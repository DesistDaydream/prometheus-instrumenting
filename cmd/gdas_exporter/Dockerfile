FROM golang:1.15 as builder
WORKDIR /root/prometheus-instrumenting
COPY go.mod go.sum ./
COPY exporter exporter
ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,https://goproxy.io,direct
RUN go build ./cmd/gdas_exporter/*.go

FROM alpine
WORKDIR /root/prometheus-instrumenting
COPY --from=builder /root/prometheus-instrumenting/main /usr/local/bin/gdas-exporter
EXPOSE 8080
ENTRYPOINT  [ "/usr/local/bin/gdas-exporter" ]
