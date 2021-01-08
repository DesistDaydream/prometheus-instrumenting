FROM golang:1.15 as builder
WORKDIR /root/exporter
COPY go.mod go.sum ./
COPY xsky_exporter xsky_exporter
ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,https://goproxy.io,direct
RUN go build ./xsky_exporter/*.go

FROM alpine
WORKDIR /root/exporter
COPY --from=builder /root/exporter/main /usr/local/bin/xsky-exporter
EXPOSE 8080
ENTRYPOINT  [ "/usr/local/bin/xsky-exporter" ]
