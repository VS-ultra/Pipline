package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// RingBuffer структура кольцевого буфера

type RingBuffer struct {
	inputChannel  chan int
	outputChannel chan int
	buffer        []int
	size          int
}

// NewRingBuffer создает новый кольцевой буфер заданного размера

func NewRingBuffer(size int) *RingBuffer {
	if size <= 0 {
		return nil
	}
	input := make(chan int)
	output := make(chan int, size)
	buffer := make([]int, size)
	return &RingBuffer{
		inputChannel:  input,
		outputChannel: output,
		buffer:        buffer,
		size:          size,
	}
}

// Run запускает кольцевой буфер

func (rb *RingBuffer) Run(done <-chan bool, interval time.Duration) {
	var start, end int
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case data := <-rb.inputChannel:
			rb.buffer[end] = data
			end = (end + 1) % rb.size
			if end == start {
				start = (start + 1) % rb.size
			}
		case <-ticker.C:
			if start != end {
				rb.outputChannel <- rb.buffer[start]
				start = (start + 1) % rb.size
			}
		case <-done:
			close(rb.outputChannel)
			return
		}
	}
}

// Write пишет данные в кольцевой буфер

func (rb *RingBuffer) Write(data int) {
	rb.inputChannel <- data
}

// ReadConsole читает данные из консоли и отправляет их в следующую стадию

func ReadConsole(nextStage chan<- int, done chan bool) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		data := scanner.Text()
		if strings.EqualFold(data, "exit") {
			fmt.Println("Программа завершила работу!")
			close(done)
			return
		}
		num, err := strconv.Atoi(data)
		if err != nil {
			fmt.Println("Программа обрабатывает только целые числа")
			continue
		}
		nextStage <- num
	}
}

// NegFilter фильтрует отрицательные числа

func NegFilter(previousStageChan <-chan int, nextStageChan chan<- int, done chan bool) {
	for {
		select {
		case data := <-previousStageChan:
			if data > 0 {
				nextStageChan <- data
			}
		case <-done:
			close(nextStageChan)
			return
		}
	}
}

// FilterThreeDiv фильтрует числа, не кратные 3 и исключает 0

func FilterThreeDiv(previousStageChan <-chan int, nextStageChan chan<- int, done chan bool) {
	for {
		select {
		case data := <-previousStageChan:
			if data != 0 && data%3 == 0 {
				nextStageChan <- data
			}
		case <-done:
			close(nextStageChan)
			return
		}
	}
}

func main() {
	var bufferSize int
	var bufferIntervalSec int

	fmt.Print("Введите размер буфера: ")
	fmt.Scan(&bufferSize)
	fmt.Print("Введите интервал буфера в секундах: ")
	fmt.Scan(&bufferIntervalSec)

	done := make(chan bool)
	input := make(chan int)
	negFilterChannel := make(chan int)
	divThreeFilterChannel := make(chan int)

	rb := NewRingBuffer(bufferSize)
	go rb.Run(done, time.Duration(bufferIntervalSec)*time.Second)
	go ReadConsole(input, done)
	go NegFilter(input, negFilterChannel, done)
	go FilterThreeDiv(negFilterChannel, divThreeFilterChannel, done)

	// Потребитель данных из кольцевого буфера
	go func() {
		for data := range divThreeFilterChannel {
			rb.Write(data)
		}
	}()

	// Вывод обработанных данных
	go func() {
		for data := range rb.outputChannel {
			fmt.Printf("Получены данные: %d\n", data)
		}
	}()

	<-done
}
