package main

import (
	"log"
	"net/http"

	"github.com/DesistDaydream/exporter/practice/harbor_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
)

func main() {
	var h collector.HarborConnInfo
	// 加载关于 Harbor 相关的 Flags
	h.HarborConnFlags()
	pflag.Parse()
	collector.HarborHealthCollector(&h)

	// 为所有 metrics 注册一个 Collector
	// 为 HarborHealthStatus 这个 metrics 注册一个 Collector
	prometheus.MustRegister(collector.HarborHealthStatus)

	// 启动 Exporter
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
