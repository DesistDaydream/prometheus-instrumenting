package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"

	"github.com/prometheus/exporter-toolkit/web"
)

// HelloWorldMetrics 用来保存所有 Metrics，实现了 prometheus.Collector
type HelloWorldMetrics struct {
	HelloWorldDesc *prometheus.Desc
	mutex          sync.Mutex // 加锁用，与 exporter 的主要运行逻辑无关
}

// NewHelloWorldMetrics 实例化 HelloWorldMetrics，就是为所有 Mestirs 设定一些基本信息
func NewHelloWorldMetrics() *HelloWorldMetrics {
	return &HelloWorldMetrics{
		HelloWorldDesc: prometheus.NewDesc(
			"a_hello_world_exporter",              // Metric 名称
			"Help Info for Hello World Exporter ", // Metric 的帮助信息
			[]string{"name"}, nil,                 // Metric 的可变标签值的标签 与 不可变标签值的标签
		),
	}
}

// Describe 让 HelloWorldMetrics 实现 Collector 接口。将 Metrics 的描述符传到 channel 中
func (ms *HelloWorldMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- ms.HelloWorldDesc
}

// Collect 让 HelloWorldMetrics 实现 Collector 接口。采集 Metrics 的具体行为并设置 Metrics 的值类型,将 Metrics 的信息传到 channel 中
func (ms *HelloWorldMetrics) Collect(ch chan<- prometheus.Metric) {
	ms.mutex.Lock() // 加锁
	defer ms.mutex.Unlock()

	ch <- prometheus.MustNewConstMetric(ms.HelloWorldDesc, prometheus.GaugeValue, float64(rand.Int31n(1000)),
		"haohao",
	)
	ch <- prometheus.MustNewConstMetric(ms.HelloWorldDesc, prometheus.GaugeValue, float64(rand.Int31n(100)),
		"nana",
	)
}

func main() {
	// Prometheus 的 客户端库 提供了一个用于日志功能的库
	promslogConfig := &promslog.Config{}
	logger := promslog.New(promslogConfig)

	m := NewHelloWorldMetrics()
	reg := prometheus.NewRegistry()
	reg.MustRegister(m)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	srv := &http.Server{}

	addr := []string{":8080"}
	isSocket := false
	configFile := "config/web-config.yaml"
	// 利用 exporter-toolkit 实现 web 能力
	// prometheus/exporter-toolkit 项目是一个工具包，可以帮助开发者实现基本的 Web 能力。
	// e.g. HTTPS, Basic Auth, etc. 都封装在 exporter-toolkit 可以直接调用
	webFlags := &web.FlagConfig{
		WebListenAddresses: &addr,
		WebSystemdSocket:   &isSocket,
		WebConfigFile:      &configFile,
	}

	if err := web.ListenAndServe(srv, webFlags, logger); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
