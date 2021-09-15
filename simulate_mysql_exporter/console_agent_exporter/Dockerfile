FROM golang:1.15 as builder
WORKDIR /root/exporter
COPY ./ ./
ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,https://goproxy.io,direct
RUN go build ./simulate_mysql_exporter/console_agent_exporter/*.go

FROM alpine
WORKDIR /root/exporter
COPY --from=builder /root/exporter/main /usr/local/bin/consoler-exporter
ENTRYPOINT  [ "/usr/local/bin/consoler-exporter" ]