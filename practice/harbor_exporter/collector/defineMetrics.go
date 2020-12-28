package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HarborHealthStatus 是 Habor 各组件健康状态的指标
	HarborHealthStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "harbor_components_health_status",
		Help: "Harbor overall health of all components",
	},
		[]string{"component"},
	)
)

// RegisterMetrics 注册所有自定的 Metrics
func RegisterMetrics() {
	// 为所有 metrics 注册一个 Collector
	prometheus.MustRegister(HarborHealthStatus)
}
