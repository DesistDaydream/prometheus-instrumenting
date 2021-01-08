package collector

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/DesistDaydream/exporter/simulate_mysql_exporter/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// check interface
	_ scraper.CommonScraper = ScrapeCluster{}

	// 设置 Metric 的基本信息，从 xsky 的接口中获取 cluster 相关的数据。
	// 由于 cluster 中包含大量内容，如果在抓取 Metrics 时，想要获取其中的所有数据
	// 则可以将 cluster 的json格式的响应体中的 key 当作 metric 的标签值即可
	cluster = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_info"),
		"Xsky Cluster Info",
		[]string{"comments"}, nil,
	)
)

// ScrapeCluster 是将要实现 Scraper 接口的一个 Metric 结构体
type ScrapeCluster struct{}

// Name 指定自己定义的 抓取器 的名字，与 Metric 的名字不是一个概念，但是一般保持一致
// 该方法用于为 ScrapeCluster 结构体实现 Scraper 接口
func (ScrapeCluster) Name() string {
	return "cluster_info"
}

// Help 指定自己定义的 抓取器 的帮助信息，这里的 Help 的内容将会作为命令行标志的帮助信息。与 Metric 的 Help 不是一个概念。
// 该方法用于为 ScrapeCluster 结构体实现 Scraper 接口
func (ScrapeCluster) Help() string {
	return "Xsky Cluster Info"
}

// Scrape 从客户端采集数据，并将其作为 Metric 通过 channel(通道) 发送。主要就是采集 Xsky 集群信息的具体行为。
// 该方法用于为 ScrapeCluster 结构体实现 Scraper 接口
func (ScrapeCluster) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	// 声明需要绑定的 响应体 与 结构体
	var (
		respBody []byte
		data     clusterJSON
	)

	// 根据 URI 获取 Response Body，获取 cluster 相关的信息。里面包含大量内容
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
	// cluster 中各种数据的 key 可以作为 metric 的标签值，cluster 中数据的值，就是该 metric 的值
	ch <- prometheus.MustNewConstMetric(cluster, prometheus.GaugeValue, float64(data.Cluster.Samples[0].UsedKbyte), "used_kbyte")
	ch <- prometheus.MustNewConstMetric(cluster, prometheus.GaugeValue, float64(data.Cluster.Samples[0].ActualKbyte), "actual_kbyte")
	return nil
}

// clusterJSON 存储 Xsky Cluster 相关信息的 Response Body 的数据
type clusterJSON struct {
	Cluster Cluster `json:"cluster"`
}

// Cluster 是 clusterJSON 的子集
type Cluster struct {
	AccessToken          string    `json:"access_token"`
	AccessURL            string    `json:"access_url"`
	Create               time.Time `json:"create"`
	DiskLightingMode     string    `json:"disk_lighting_mode"`
	DownOutInterval      int       `json:"down_out_interval"`
	ElasticsearchEnabled bool      `json:"elasticsearch_enabled"`
	FsID                 string    `json:"fs_id"`
	ID                   int       `json:"id"`
	Maintained           bool      `json:"maintained"`
	Name                 string    `json:"name"`
	OsGatewayOplogSwitch bool      `json:"os_gateway_oplog_switch"`
	Samples              []Samples `json:"samples"`
	SnmpEnabled          bool      `json:"snmp_enabled"`
	StatsReservedDays    int       `json:"stats_reserved_days"`
	Status               string    `json:"status"`
	Update               time.Time `json:"update"`
	Version              string    `json:"version"`
}

// Samples 是 Cluster 的子集，一个数组
type Samples struct {
	ActualKbyte            int64     `json:"actual_kbyte"`
	Create                 time.Time `json:"create"`
	DataKbyte              int64     `json:"data_kbyte"`
	DegradedPercent        int       `json:"degraded_percent"`
	ErrorKbyte             int       `json:"error_kbyte"`
	HealthyPercent         int       `json:"healthy_percent"`
	OsDownBandwidthKbyte   int       `json:"os_down_bandwidth_kbyte"`
	OsDownIops             int       `json:"os_down_iops"`
	OsMergeSpeed           int       `json:"os_merge_speed"`
	OsUpBandwidthKbyte     int       `json:"os_up_bandwidth_kbyte"`
	OsUpIops               int       `json:"os_up_iops"`
	ReadBandwidthKbyte     int       `json:"read_bandwidth_kbyte"`
	ReadIops               int       `json:"read_iops"`
	ReadLatencyUs          int       `json:"read_latency_us"`
	RecoveryBandwidthKbyte int       `json:"recovery_bandwidth_kbyte"`
	RecoveryIops           int       `json:"recovery_iops"`
	RecoveryPercent        int       `json:"recovery_percent"`
	TotalKbyte             int64     `json:"total_kbyte"`
	UnavailablePercent     int       `json:"unavailable_percent"`
	UsedKbyte              int64     `json:"used_kbyte"`
	WriteBandwidthKbyte    int       `json:"write_bandwidth_kbyte"`
	WriteIops              int       `json:"write_iops"`
	WriteLatencyUs         int       `json:"write_latency_us"`
}
