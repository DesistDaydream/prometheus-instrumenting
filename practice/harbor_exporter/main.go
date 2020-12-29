package main

import (
	"net/http"

	"github.com/DesistDaydream/exporter/practice/harbor_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	collector.Conn()
	// 注册所有自定义的 Metrics
	n := collector.NewHarborMetrics()
	reg := prometheus.NewRegistry()
	reg.MustRegister(n)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	// 启动 Exporter
	// http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
