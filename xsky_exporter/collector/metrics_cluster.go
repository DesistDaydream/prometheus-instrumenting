package collector

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// check interface
	_ Scraper = ScrapeCluster{}

	// 设置 Metric 的基本信息
	cluster = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_info"),
		"Xsky Cluster Info",
		[]string{"comments"}, nil,
	)
)

// ScrapeCluster 将要实现 Scraper 接口的一个 Metric 结构体
type ScrapeCluster struct{}

// Name of the Scraper. Should be unique.
func (ScrapeCluster) Name() string {
	return "cluster_info"
}

// Help describes the role of the Scraper.
func (ScrapeCluster) Help() string {
	return "Xsky Cluster Info"
}

// Scrape collects data from client and sends it over channel as prometheus metric.
func (ScrapeCluster) Scrape(client *XskyClient, ch chan<- prometheus.Metric) (err error) {
	// 声明需要绑定的 响应体 与 结构体
	var (
		respBody []byte
		data     clusterJSON
	)

	// 根据 URI 获取 Response Body
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