FROM golang:1.16-alpine as builder
WORKDIR /root/exporter

ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOPROXY=https://goproxy.cn,https://goproxy.io,https://mirrors.aliyun.com/goproxy/,direct
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY ./ /root/exporter
RUN go build -o e37-exporter ./e37_exporter/*.go

FROM alpine
WORKDIR /root/exporter
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add --no-cache tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
ENV TZ=Asia/Shanghai
COPY --from=builder /root/exporter/e37-exporter /usr/local/bin/e37-exporter
ENTRYPOINT  [ "/usr/local/bin/e37-exporter" ]