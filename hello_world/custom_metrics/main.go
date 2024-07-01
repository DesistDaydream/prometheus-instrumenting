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
	cpuTemp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "a_cpu_temperature_celsius",
		Help: "Current temperature of the CPU.",
	})
	hdFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "a_hd_errors_total",
			Help: "Number of hard-disk errors.",
		},
		[]string{"device"},
	)
)

// 注册那些自定义的 Metrics，以便可以暴露这些 Metrics。
func init() {
	prometheus.MustRegister(cpuTemp, hdFailures)
}

func main() {
	// prometheus 提供了多种方法，可以为 Metrics 设置值和标签。
	// 使用随机数为 cpuTemp 指标设置一个值
	//
	// TODO：为啥 Metric 的值每次 GET 不会变？~迷茫ing....
	//
	cpuTemp.Set(float64(rand.Int31n(100)))
	// 为 hdFailures 指标的 device:/dev/sda 标签设置一个值
	hdFailures.With(prometheus.Labels{"device": "/dev/sda"}).Inc()

	// Handler 函数提供了默认的处理器，以便通过 HTTP 服务器暴露注册过的 metrics。一般情况下，使用 "/metrics" 路径来暴露 metrics
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
