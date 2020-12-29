package collector

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HarborHealthStatus 是 Habor 各组件健康状态的指标
	HarborHealthStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "harbor_components_health_status",
			Help: "Harbor overall health of all components",
		},
		[]string{"component"},
	)
)

// HarborMetrics 应该采集的 Harbor 的 metrics
type HarborMetrics struct {
	HarborHealthStatus *prometheus.Desc
	mutex              sync.Mutex
}

// NewHarborMetrics 实例化 Metrics
func NewHarborMetrics() *HarborMetrics {
	return &HarborMetrics{
		HarborHealthStatus: prometheus.NewDesc(
			"harbor_components_health_status",
			"Harbor overall health of all components",
			[]string{"component"}, nil,
		),
	}
}

// Describe 让 HelloWorldMetrics 实现 Collector 接口。将 Metrics 的描述符传到 channel 中
func (ms *HarborMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- ms.HarborHealthStatus
}

// Collect 让 HelloWorldMetrics 实现 Collector 接口。采集 Metrics 的具体行为并设置 Metrics 的值类型,将 Metrics 的信息传到 channel 中
func (ms *HarborMetrics) Collect(ch chan<- prometheus.Metric) {
	ms.mutex.Lock() // 加锁
	defer ms.mutex.Unlock()

	// 采集 HarborHealthStatus 指标
	var hcs HarborComponentStatue
	hcs.HarborHealthCollector(&HC)
	// fmt.Println(hcs)
	for i, data := range hcs.ComponentNames {
		fmt.Println(data)
		ch <- prometheus.MustNewConstMetric(ms.HarborHealthStatus, prometheus.GaugeValue, hcs.ComponentStatus[i], hcs.ComponentNames[i])
	}
}
