package collector

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// 在这里定义 Metrics 以及实现 prometheus.Collector 接口。
// 想要采集多个指标，只需要在本代码的各个 XXXXX 部分添加即可。具体的采集实现，放在其他文件中。

// HarborMetrics 应该采集的 Harbor 的 metrics
type HarborMetrics struct {
	HarborHealthStatus *prometheus.Desc
	XXXX               *prometheus.Desc
	mutex              sync.Mutex
}

// NewHarborMetrics 实例化 Metrics
func NewHarborMetrics() *HarborMetrics {
	return &HarborMetrics{
		// 指标之一，Harbor 各组件健康状态
		HarborHealthStatus: prometheus.NewDesc(
			"harbor_components_health_status",
			"Harbor overall health of all components",
			[]string{"component"}, nil,
		),
		// 指标之一，
		XXXX: prometheus.NewDesc(
			"harbor_XXXXX",
			"Harbor XXXXX",
			nil, nil,
		),
	}
}

// Describe 让 HelloWorldMetrics 实现 Collector 接口。将 Metrics 的描述符传到 channel 中
func (ms *HarborMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- ms.HarborHealthStatus
	ch <- ms.XXXX
}

// Collect 让 HelloWorldMetrics 实现 Collector 接口。采集 Metrics 的具体行为并设置 Metrics 的值类型,将 Metrics 的信息传到 channel 中
func (ms *HarborMetrics) Collect(ch chan<- prometheus.Metric) {
	ms.mutex.Lock() // 加锁
	defer ms.mutex.Unlock()

	// 采集 HarborHealthStatus 指标
	h := HarborHealthCollector(&HC)
	for i, componentName := range h.ComponentNames {
		ch <- prometheus.MustNewConstMetric(ms.HarborHealthStatus, prometheus.GaugeValue, h.ComponentStatus[i], componentName)
	}

	// 采集 XXX 指标
	ch <- prometheus.MustNewConstMetric(ms.XXXX, prometheus.GaugeValue, 100)
}
