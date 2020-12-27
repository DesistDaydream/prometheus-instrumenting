package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HarborHealthStatus 是 Habor 各组件健康状态的指标
	HarborHealthStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "harbor_components_health_status",
		Help: "Harbor overall health of all components",
	})
)

// RegisterMetrics 注册所有自定的 Metrics
func RegisterMetrics() {
	// 为所有 metrics 注册一个 Collector
	// 为 HarborHealthStatus 这个 metrics 注册一个 Collector
	prometheus.MustRegister(HarborHealthStatus)
}
