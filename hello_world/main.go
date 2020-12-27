package main

import (
	"log"
	"math/rand"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics 用来保存所有 Metrics
type Metrics struct {
	MetricsDesc map[string]*prometheus.Desc
	mutex       sync.Mutex
}

// NewMetrics 实例化所有的 Metrics
func NewMetrics() *Metrics {
	return &Metrics{
		MetricsDesc: map[string]*prometheus.Desc{
			// 定义 Metrics 信息，名称、帮助信息、标签 等等
			"exporter_hello_world": prometheus.NewDesc("exporter_hello_world", "Help Info for exporter hello world", []string{"name"}, nil),
		},
	}
}

// Describe 让 Metrics 实现 Collector 接口。获取所有 Metrics 的描述符
func (c *Metrics) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.MetricsDesc {
		ch <- m
	}
}

// Collect 让 Metrics 实现 Collector 接口。采集 Metrics 的具体行为
func (c *Metrics) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock() // 加锁
	defer c.mutex.Unlock()

	// 获取 Metrics 的值
	data := c.Set()
	// 为 Metric 设置类型，并将获取到的值循环输出。如果由多个 Metircs 则写多个 for 循环即可
	for name, currentValue := range data {
		ch <- prometheus.MustNewConstMetric(c.MetricsDesc["exporter_hello_world"], prometheus.CounterValue, float64(currentValue), name)
	}
}

// Set 为 Metric 设置值。如果由多个 Metrics，则可以设置多个返回值
func (c *Metrics) Set() (value map[string]int) {
	value = map[string]int{
		// 使用 rand 生成随机数
		"desistdaydream.com": int(rand.Int31n(1000)),
		"nana.com":           int(rand.Int31n(1000)),
	}
	return
}

func main() {
	// 实例化自己定义的所有 Metrics
	m := NewMetrics()
	// 实例化一个默认的注册器
	registry := prometheus.NewRegistry()
	// 注册我们自己定义的所有 Metrics
	registry.MustRegister(m)
	// 使用 HandlerFor 将自己定义的已注册的 Metrics 作为参数，以便可以通过 http 获取 metric 信息
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
