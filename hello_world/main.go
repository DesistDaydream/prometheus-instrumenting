package main

import (
	"log"
	"math/rand"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 声明 Metrics，并定义这些 Metrics 的元数据，名称、帮助信息、值类型等等
// 每个 Metrics 名称前加了 a，主要是为了在获取 Metrics 的时候，这些自定义的 Metric 可以排在前面
var (
	cpuTemp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "z_cpu_temperature_celsius",
			Help: "Current temperature of the CPU.",
		})
	hdFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "z_hd_errors_total",
			Help: "Number of hard-disk errors.",
		},
		[]string{"device"},
	)
)

func main() {
	// 注册那些自定义的 Metrics，以便可以暴露这些 Metrics。
	prometheus.MustRegister(cpuTemp, hdFailures)
	// prometheus 提供了多种方法，可以为 Metrics 设置值和标签。
	cpuTemp.Set(float64(rand.Int31n(100))) // TODO：为啥 Metric 的值每次 GET 不会变？~迷茫ing....
	// 为 hdFailures 指标的 device:/dev/sda 标签设置一个值
	hdFailures.With(prometheus.Labels{"device": "/dev/sda"}).Inc()

	// promhttp.Handler() 函数旨再涵盖大部分基本用例，提供了默认的 Metrics 处理器。
	// 如果需要更多自定义的操作（包括使用非默认 Gatherer、不同的检测和非默认 HandlerOpts），使用 HandlerFor() 函数。
	// 常用的处理器有如下几个
	// 1. Handler()
	// 2. HandlerFor()
	// 3. HandlerForTransactional()
	// 4. InstrumentMetricHandler()
	// TODO: 整理这四个的异同。很多时候（e.g. node-exporter、mysql-exporter 都不是直接使用 Handler 或者 HandlerFor）
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
