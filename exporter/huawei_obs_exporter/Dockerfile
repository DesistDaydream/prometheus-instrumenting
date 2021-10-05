FROM golang:1.15 as builder
WORKDIR /root/exporter

ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOPROXY=https://goproxy.cn,https://goproxy.io,https://mirrors.aliyun.com/goproxy/,direct
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY ./ /root/exporter
RUN go build -o huawei-obs-exporter ./huawei_obs_exporter/*.go

FROM alpine
WORKDIR /root/exporter
COPY --from=builder /root/exporter/huawei-obs-exporter /usr/local/bin/huawei-obs-exporter
ENTRYPOINT  [ "/usr/local/bin/huawei-obs-exporter" ]