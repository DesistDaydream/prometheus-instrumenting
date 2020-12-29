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
	// 实例化 Metric 获取他们的 Desc
	n := collector.NewHarborMetrics()
	// 实例化一个新注册器
	reg := prometheus.NewRegistry()
	// 使用新注册器注册自定义的 Metric
	reg.MustRegister(n)

	// 启动 Exporter
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.ListenAndServe(":8080", nil)
}
