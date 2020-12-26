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
