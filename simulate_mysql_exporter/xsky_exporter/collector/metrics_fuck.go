package collector

import (
	"encoding/json"
	"fmt"

	"github.com/DesistDaydream/exporter/simulate_mysql_exporter/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// check interface
	_ scraper.CommonScraper = ScrapeFuck{}

	// 设置 Metric 的基本信息，从 xsky 的接口中获取 fuck 相关的数据。
	// 由于 fuck 中包含大量内容，如果在抓取 Metrics 时，想要获取其中的所有数据
	// 则可以将 fuck 的json格式的响应体中的 key 当作 metric 的标签值即可
	fuck = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "fuck_info"),
		"Xsky Cluster Info",
		[]string{"comments"}, nil,
	)
)

// ScrapeFuck 是将要实现 Scraper 接口的一个 Metric 结构体
type ScrapeFuck struct{}

// Name 指定自己定义的 抓取器 的名字，与 Metric 的名字不是一个概念，但是一般保持一致
// 该方法用于为 ScrapeFuck 结构体实现 Scraper 接口
func (ScrapeFuck) Name() string {
	return "fuck_info"
}

// Help 指定自己定义的 抓取器 的帮助信息，这里的 Help 的内容将会作为命令行标志的帮助信息。与 Metric 的 Help 不是一个概念。
// 该方法用于为 ScrapeFuck 结构体实现 Scraper 接口
func (ScrapeFuck) Help() string {
	return "Xsky Cluster Info"
}

// Scrape 从客户端采集数据，并将其作为 Metric 通过 channel(通道) 发送。主要就是采集 Xsky 集群信息的具体行为。
// 该方法用于为 ScrapeFuck 结构体实现 Scraper 接口
func (ScrapeFuck) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	// 声明需要绑定的 响应体 与 结构体
	var (
		respBody []byte
		data     clusterJSON
	)

	// 根据 URI 获取 Response Body，获取 fuck 相关的信息。里面包含大量内容
	url := "/v1/cluster"
	if respBody, err = client.Request(url); err != nil {
		return err
	}

	// 绑定 Body 与 struct
	if err = json.Unmarshal(respBody, &data); err != nil {
		return err
	}

	// 根据 Response Body 获取用户使用量
	fmt.Printf("当前用户已经使用了 %v KiB\n", data.Cluster.Samples[0].UsedKbyte)
	// fuck 中各种数据的 key 可以作为 metric 的标签值，fuck 中数据的值，就是该 metric 的值
	ch <- prometheus.MustNewConstMetric(fuck, prometheus.GaugeValue, float64(data.Cluster.Samples[0].UsedKbyte), "used_kbyte")
	ch <- prometheus.MustNewConstMetric(fuck, prometheus.GaugeValue, float64(data.Cluster.Samples[0].ActualKbyte), "actual_kbyte")
	return nil
}
