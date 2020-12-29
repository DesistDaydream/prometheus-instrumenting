package main

import (
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
			// 定义一个 Metrics 信息，名称、帮助信息、标签 等等
			"exporter_hello_world": prometheus.NewDesc(
				"exporter_hello_world",               // Metric 名称
				"Help Info for exporter hello world", // Metric 的帮助信息
				[]string{"name"},                     // Metric 的可变标签值的标签
				nil,                                  // Metric 的不可变标签值的标签
			),
			// 还可以定义第二个、第三个.... Metrics
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
	// 为 Metric 及其 标签 赋予具体的值
	// 如果有多个 Metircs 则写多个 for 循环即可
	for name, currentValue := range data {
		// MustNewConstMetric() 返回一个具有固定值且无法更改的 Metric，需要传递的四个参数分别为：
		// Metric 的 Desc、Metric 的值类型、当前 Metric 的值、Metric 的标签的值...(可以有多个标签值)
		ch <- prometheus.MustNewConstMetric(c.MetricsDesc["exporter_hello_world"], prometheus.CounterValue, float64(currentValue), name)
	}
}

// Set 为 Metric 设置值。如果由多个 Metrics，则可以设置多个返回值
func (c *Metrics) Set() (value map[string]int) {
	value = map[string]int{
		// desistdaydream 和 nana 是这个 Metric 的 name 标签的值
		// 使用 rand 生成的随机数，是具有不同标签值的 Metric 的值
		"desistdaydream": int(rand.Int31n(1000)),
		"nana":           int(rand.Int31n(1000)),
	}
	return
}

func main() {
	// 实例化自己定义的所有 Metrics
	m := NewMetrics()

	// 两种注册 Metrics 的方式
	//
	// 第一种：实例化一个新注册器，用来注册 自定义Metrics
	// 使用 HandlerFor 将自己定义的已注册的 Metrics 作为参数，以便可以通过 http 获取 metric 信息
	// reg := prometheus.NewRegistry()
	// reg.MustRegister(m)
	// http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	//
	// 第二种：使用自带的默认注册器，用来注册 自定义Metrics
	prometheus.MustRegister(m)
	http.Handle("/metrics", promhttp.Handler())

	// 让该 exporter 监听在8080 上
	http.ListenAndServe(":8080", nil)
}

/*
Export 暴露结果：
# HELP exporter_hello_world Help Info for exporter hello world
# TYPE exporter_hello_world counter
exporter_hello_world{name="desistdaydream"} 81
exporter_hello_world{name="nana"} 887
*/
