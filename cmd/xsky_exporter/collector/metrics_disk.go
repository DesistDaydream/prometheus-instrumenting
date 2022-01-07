package collector

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/DesistDaydream/prometheus-instrumenting/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// check interface
	_ scraper.CommonScraper = ScrapeDisk{}

	// 设置 Metric 的基本信息，从 xsky 的接口中获取 disk 相关的数据。
	// 由于 disk 中包含大量内容，如果在抓取 Metrics 时，想要获取其中的所有数据
	// 则可以将 disk 的json格式的响应体中的 key 当作 metric 的标签值即可
	diskStatus = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "disk_status"),
		"Xsky Cluster Info",
		[]string{"disk_id", "host_name"}, nil,
	)
	diskCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "disk_count"),
		"Xsky Cluster Info",
		nil, nil,
	)
)

// ScrapeDisk 是将要实现 Scraper 接口的一个 Metric 结构体
type ScrapeDisk struct{}

// Name 指定自己定义的 抓取器 的名字，与 Metric 的名字不是一个概念，但是一般保持一致
// 该方法用于为 ScrapeDisk 结构体实现 Scraper 接口
func (ScrapeDisk) Name() string {
	return "disk_info"
}

// Help 指定自己定义的 抓取器 的帮助信息，这里的 Help 的内容将会作为命令行标志的帮助信息。与 Metric 的 Help 不是一个概念。
// 该方法用于为 ScrapeDisk 结构体实现 Scraper 接口
func (ScrapeDisk) Help() string {
	return "Xsky Cluster Info"
}

// Scrape 从客户端采集数据，并将其作为 Metric 通过 channel(通道) 发送。主要就是采集 Xsky 集群信息的具体行为。
// 该方法用于为 ScrapeDisk 结构体实现 Scraper 接口
func (ScrapeDisk) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	// 声明需要绑定的 响应体 与 结构体
	var (
		respBody []byte
		data     DisksJSON
		status   float64
	)

	// 根据 URI 获取 Response Body，获取 disk 相关的信息。里面包含大量内容
	url := "/api/v1/disks"
	if respBody, err = client.Request("GET", url, nil); err != nil {
		return err
	}

	// 绑定 Body 与 struct
	if err = json.Unmarshal(respBody, &data); err != nil {
		return err
	}

	// 根据 Response Body 获取用户使用量
	fmt.Printf("当前一共有 %v 块磁盘\n", data.Paging.TotalCount)
	// disk 中各种数据的 key 可以作为 metric 的标签值，disk 中数据的值，就是该 metric 的值
	ch <- prometheus.MustNewConstMetric(diskCount, prometheus.GaugeValue, float64(data.Paging.TotalCount))
	for i := 0; i < len(data.Disks); i++ {
		if data.Disks[i].ActionStatus == "active" {
			status = 1
		}
		ch <- prometheus.MustNewConstMetric(diskStatus, prometheus.GaugeValue, status, strconv.Itoa(data.Disks[i].ID), data.Disks[i].Host.Name)
	}
	return nil
}

// DisksJSON is
type DisksJSON struct {
	Disks  []Disks `json:"disks"`
	Paging Paging  `json:"paging"`
}

// Disks is
type Disks struct {
	ActionStatus   string        `json:"action_status"`
	Bytes          int64         `json:"bytes"`
	CacheCreate    time.Time     `json:"cache_create"`
	ChannelID      string        `json:"channel_id"`
	Create         time.Time     `json:"create"`
	Device         string        `json:"device"`
	DiskType       string        `json:"disk_type"`
	DriverType     string        `json:"driver_type"`
	EnclosureID    string        `json:"enclosure_id"`
	Host           Host          `json:"host"`
	ID             int           `json:"id"`
	IsCache        bool          `json:"is_cache"`
	IsRoot         bool          `json:"is_root"`
	LightingStatus string        `json:"lighting_status"`
	Model          string        `json:"model"`
	PartitionNum   int           `json:"partition_num"`
	Partitions     []interface{} `json:"partitions"`
	PowerSafe      bool          `json:"power_safe"`
	RotationRate   string        `json:"rotation_rate"`
	Rotational     bool          `json:"rotational"`
	Samples        []DiskSamples `json:"samples"`
	Serial         string        `json:"serial"`
	SlotID         string        `json:"slot_id"`
	SmartAttrs     []interface{} `json:"smart_attrs"`
	SsdLifeLeft    interface{}   `json:"ssd_life_left"`
	Status         string        `json:"status"`
	Update         time.Time     `json:"update"`
	Used           bool          `json:"used"`
	Wwid           string        `json:"wwid"`
}

// Host is
type Host struct {
	AdminIP string `json:"admin_ip"`
	ID      int    `json:"id"`
	Name    string `json:"name"`
}

// DiskSamples is
type DiskSamples struct {
	AvgQueueLen         int       `json:"avg_queue_len"`
	Create              time.Time `json:"create"`
	DegradedPercent     int       `json:"degraded_percent"`
	HealthyPercent      int       `json:"healthy_percent"`
	IoUtil              float64   `json:"io_util"`
	KbytePerIo          int       `json:"kbyte_per_io"`
	OmapTotalKbyte      int       `json:"omap_total_kbyte"`
	OmapUsedKbyte       int       `json:"omap_used_kbyte"`
	OmapUsedPercent     float64   `json:"omap_used_percent"`
	ReadBandwidthKbyte  int       `json:"read_bandwidth_kbyte"`
	ReadIops            int       `json:"read_iops"`
	ReadMergedPs        int       `json:"read_merged_ps"`
	ReadWaitUs          int       `json:"read_wait_us"`
	RecoveryPercent     int       `json:"recovery_percent"`
	TotalBandwidthKbyte int       `json:"total_bandwidth_kbyte"`
	TotalIoWaitUs       int       `json:"total_io_wait_us"`
	TotalIops           int       `json:"total_iops"`
	TotalKbyte          int64     `json:"total_kbyte"`
	UnavailablePercent  int       `json:"unavailable_percent"`
	UsedKbyte           int       `json:"used_kbyte"`
	UsedPercent         float64   `json:"used_percent"`
	WriteBandwidthKbyte int       `json:"write_bandwidth_kbyte"`
	WriteIops           int       `json:"write_iops"`
	WriteMergedPs       int       `json:"write_merged_ps"`
	WriteWaitUs         int       `json:"write_wait_us"`
}

// Paging is
type Paging struct {
	Count      int `json:"count"`
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
	TotalCount int `json:"total_count"`
}
