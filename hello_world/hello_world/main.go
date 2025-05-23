package main

import (
	"math/rand"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HelloWorldMetrics 用来保存所有 Metrics,实现了 prometheus.Collector
type HelloWorldMetrics struct {
	HelloWorldPersonDesc *prometheus.Desc
	HelloWorldHomeDesc   *prometheus.Desc
	mutex                sync.Mutex
}

// NewHelloWorldMetrics 实例化 HelloWorldMetrics，就是为所有 Mestirs 设定一些基本信息
func NewHelloWorldMetrics() *HelloWorldMetrics {
	return &HelloWorldMetrics{
		HelloWorldPersonDesc: prometheus.NewDesc(
			"a_hello_world_person_exporter",              // Metric 名称
			"Help Info for Hello World Person Exporter ", // Metric 的帮助信息
			[]string{"name"}, nil,                        // Metric 的可变标签值的标签 与 不可变标签值的标签
		),
		HelloWorldHomeDesc: prometheus.NewDesc(
			"a_hello_world_home_exporter",              // Metric 名称
			"Help Info for Hello World Home Exporter ", // Metric 的帮助信息
			nil, nil, // Metric 的可变标签值的标签 与 不可变标签值的标签，这里都为 nil 表示该 Metric 没有标签
		),
	}
}

// Describe 让 HelloWorldMetrics 实现 Collector 接口。将 Metrics 的描述符传到 channel 中
func (ms *HelloWorldMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- ms.HelloWorldPersonDesc
	ch <- ms.HelloWorldHomeDesc
}

// Collect 让 HelloWorldMetrics 实现 Collector 接口。采集 Metrics 的具体行为并设置 Metrics 的值类型,将 Metrics 的信息传到 channel 中
func (ms *HelloWorldMetrics) Collect(ch chan<- prometheus.Metric) {
	ms.mutex.Lock() // 加锁
	defer ms.mutex.Unlock()

	// 为 ms.HelloWorldDesc 这个 Metric 设置其属性的值
	// 该 Metric 值的类型为 Gauge，name 标签值为 haohao 时，Metric 的值为 1000 以内的随机数
	ch <- prometheus.MustNewConstMetric(ms.HelloWorldPersonDesc, prometheus.GaugeValue, float64(rand.Int31n(1000)),
		"haohao",
	)
	// 该 Metric 值的类型为 Gauge，name 标签值为 nana 时，Metric 的值为 100 以内的随机数
	ch <- prometheus.MustNewConstMetric(ms.HelloWorldPersonDesc, prometheus.GaugeValue, float64(rand.Int31n(100)),
		"nana",
	)
	// 为 ms.HelloWorldHomeDesc 这个 Metric 设置其属性的值，该 Metric 的值，是通过外部函数调用获得的。
	// 由于该指标没有标签，所以第四个及以后的参数不写即可
	ch <- prometheus.MustNewConstMetric(ms.HelloWorldHomeDesc, prometheus.GaugeValue, SetHelloWorldHomeMetricValue())
}

// SetHelloWorldHomeMetricValue 我们可以将设置 Metrics 值的行为，拿出来单独定义，然后从 Collect 中引用这个函数即可
func SetHelloWorldHomeMetricValue() float64 {
	return float64(rand.Int31n(10000))
}

func main() {
	// 实例化自己定义的所有 Metrics
	m := NewHelloWorldMetrics()

	// 两种注册 Metrics 的方式
	//
	// 第一种：实例化一个新注册器，用来注册 自定义Metrics
	// 使用 HandlerFor 将自己定义的已注册的 Metrics 作为参数，以便可以通过 http 获取 metric 信息
	reg := prometheus.NewRegistry()
	// 可以看到 MustRegister() 需要指定 Collector 接口作为参数
	// 所以我们的 m 也必须要实现 Collecor 接口的两个方法，以便注册完成后，prometheus 库可以直接调用我们设计的这俩方法的行为进行后续操作
	reg.MustRegister(m)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	//
	// 第二种：使用自带的默认注册器，用来注册 自定义Metrics
	// prometheus.MustRegister(m)
	// http.Handle("/metrics", promhttp.Handler())

	// 让该 exporter 监听在8080 上
	http.ListenAndServe(":8080", nil)
}

/*
Export 暴露结果：
# HELP a_hello_world_exporter Help Info for Hello World Exporter
# TYPE a_hello_world_exporter counter
a_hello_world_exporter{name="DesistDaydream"} 59
*/
