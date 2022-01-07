package collector

import (
	"encoding/json"

	"github.com/DesistDaydream/prometheus-instrumenting/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ scraper.CommonScraper = ScrapeCluster{}

	clusterServerCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "cluster_server_count"),
		"集群中节点总数",
		[]string{}, nil,
	)

	clusterServerStatus = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "cluster_server_status"),
		"集群中节点状态",
		[]string{"name", "serial_number", "management_ip"}, nil,
	)
)

// ScrapeCluster 是将要实现 Scraper 接口的一个 Metric 结构体
type ScrapeCluster struct{}

// Name 指定自己定义的 抓取器 的名字，与 Metric 的名字不是一个概念，但是一般保持一致
func (ScrapeCluster) Name() string {
	return "cluster_server_info"
}

// Help 指定自己定义的 抓取器 的帮助信息，这里的 Help 的内容将会作为命令行标志的帮助信息。与 Metric 的 Help 不是一个概念。
func (ScrapeCluster) Help() string {
	return "HWObs Cluster Server info"
}

// Scrape 从客户端采集数据，并将其作为 Metric 通过 channel(通道) 发送。主要就是采集 HWObs 集群信息的具体行为。
func (ScrapeCluster) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	var (
		respBody          []byte
		clusterServerData clusterServerData
	)

	url := "/api/v2/cluster/servers"
	if respBody, err = client.Request("GET", url, nil); err != nil {
		return err
	}

	if err = json.Unmarshal(respBody, &clusterServerData); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(clusterServerCount, prometheus.GaugeValue, float64(len(clusterServerData.Data)))

	for i := 0; i < len(clusterServerData.Data); i++ {
		ch <- prometheus.MustNewConstMetric(clusterServerStatus, prometheus.GaugeValue, float64(clusterServerData.Data[i].Status),
			clusterServerData.Data[i].Name,
			clusterServerData.Data[i].SerialNumber,
			clusterServerData.Data[i].ManagementIP,
		)
	}

	return nil
}

// ClusterServerData 存储 HWObs Cluster 相关信息的 Response Body 的数据
type clusterServerData struct {
	Data   []data `json:"data"`
	Result result `json:"result"`
}
type data struct {
	ID int `json:"id"`
	// 节点名称，即主机名
	Name string `json:"name"`
	// 节点状态，0-正常，1-异常
	Status             int         `json:"status"`
	Cabinet            string      `json:"cabinet"`
	Subrack            string      `json:"subrack"`
	SlotNumber         string      `json:"slot_number"`
	Model              string      `json:"model"`
	InCluster          bool        `json:"in_cluster"`
	InstallationStatus bool        `json:"installation_status"`
	SoftwareVersion    string      `json:"software_version"`
	AzID               interface{} `json:"az_id"`
	RunningStatus      string      `json:"running_status"`
	SubrackSn          interface{} `json:"subrack_sn"`
	BaseBoard          interface{} `json:"base_board"`
	Usage              []string    `json:"usage"`
	SerialNumber       string      `json:"serial_number"`
	// 管理IP
	ManagementIP string   `json:"management_ip"`
	Role         []string `json:"role"`
	Description  string   `json:"description"`
	Suggestion   string   `json:"suggestion"`
}
type result struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion"`
}
