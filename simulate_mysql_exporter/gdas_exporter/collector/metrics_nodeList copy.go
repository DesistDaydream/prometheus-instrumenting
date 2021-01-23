package collector

import (
	"encoding/json"
	"fmt"

	"github.com/DesistDaydream/exporter/simulate_mysql_exporter/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// check interface
	_ scraper.CommonScraper = ScrapeNodeList2{}

	// 设置 Metric 的基本信息
	nodelist2 = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "nodelist2_info"),
		"Gdas Magazines Info",
		[]string{"node_name"}, nil,
	)
)

// ScrapeNodeList2 是将要实现 Scraper 接口的一个 Metric 结构体
type ScrapeNodeList2 struct{}

// Name 指定自己定义的 抓取器 的名字，与 Metric 的名字不是一个概念，但是一般保持一致
// 该方法用于为 ScrapeNodeList2 结构体实现 Scraper 接口
func (ScrapeNodeList2) Name() string {
	return "nodelist2_info"
}

// Help 指定自己定义的 抓取器 的帮助信息，这里的 Help 的内容将会作为命令行标志的帮助信息。与 Metric 的 Help 不是一个概念。
// 该方法用于为 ScrapeNodeList2 结构体实现 Scraper 接口
func (ScrapeNodeList2) Help() string {
	return "Gdas Magazines Info"
}

// Scrape 从客户端采集数据，并将其作为 Metric 通过 channel(通道) 发送。主要就是采集 Gdas 集群信息的具体行为。
// 该方法用于为 ScrapeNodeList2 结构体实现 Scraper 接口
func (ScrapeNodeList2) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	// 声明需要绑定的 响应体 与 结构体
	var (
		respBody []byte
		data     NodeLists2
	)

	// 根据 URI 获取 Response Body
	url := "/v1/nodeList"
	if respBody, err = client.Request("GET", url, nil); err != nil {
		return err
	}

	// 绑定 Body 与 struct
	if err = json.Unmarshal(respBody, &data); err != nil {
		return err
	}

	fmt.Printf("第二个nodelist，当前共有 %v 个节点\n", len(data.NodeList))

	for i := 0; i < len(data.NodeList); i++ {
		fmt.Println("节点名称为：", data.NodeList[i].DamName)
		ch <- prometheus.MustNewConstMetric(nodelist2, prometheus.GaugeValue, data.NodeList[i].Status, data.NodeList[i].DamName)
	}

	return nil
}

// NodeLists2 is
type NodeLists2 struct {
	Result   string     `json:"result"`
	NodeList []NodeList `json:"nodeList"`
}

// NodeList2 is
type NodeList2 struct {
	IP      string  `json:"ip"`
	Status  float64 `json:"status"` // 节点状态，0 为正常，1 为异常
	DamName string  `json:"damName"`
}
