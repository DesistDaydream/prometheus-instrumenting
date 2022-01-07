FROM golang:1.16-alpine as builder
WORKDIR /root/prometheus-instrumenting

ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOPROXY=https://goproxy.cn,https://goproxy.io,https://mirrors.aliyun.com/goproxy/,direct
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY ./ /root/prometheus-instrumenting
RUN go build -o xsky-exporter ./cmd/xsky_exporter/*.go

FROM alpine
LABEL org.opencontainers.image.source https://github.com/DesistDaydream/prometheus-instrumenting
WORKDIR /root/prometheus-instrumenting
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add --no-cache tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
ENV TZ=Asia/Shanghai
COPY --from=builder /root/prometheus-instrumenting/xsky-exporter /usr/local/bin/xsky-exporter
ENTRYPOINT  [ "/usr/local/bin/xsky-exporter" ]
