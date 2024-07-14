FROM golang:latest AS compiling
RUN mkdir -p /go/src/webPipeline
WORKDIR /go/src/webPipeline
ADD main.go .
ADD go.mod .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
LABEL version="1.0.0"
LABEL maintainer="Sevachev Vlad <sevachev.vlad@mail.ru>"
WORKDIR /root/
COPY --from=compiling /go/src/webPipeline/app .
CMD ["./app"]
EXPOSE 8080 