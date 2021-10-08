FROM golang:1.16-alpine as builder
WORKDIR /root/exporter

ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOPROXY=https://goproxy.cn,https://goproxy.io,https://mirrors.aliyun.com/goproxy/,direct
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY ./ /root/exporter
RUN go build -o huawei-obs-exporter ./exporter/huawei_obs_exporter/*.go

FROM alpine
LABEL org.opencontainers.image.source https://github.com/DesistDaydream/prometheus-instrumenting
WORKDIR /root/exporter
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add --no-cache tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
ENV TZ=Asia/Shanghai
COPY --from=builder /root/exporter/huawei-obs-exporter /usr/local/bin/huawei-obs-exporter
ENTRYPOINT  [ "/usr/local/bin/huawei-obs-exporter" ]