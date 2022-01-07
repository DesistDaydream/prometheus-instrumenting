package collector

import (
	"encoding/json"

	"github.com/DesistDaydream/prometheus-instrumenting/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ scraper.CommonScraper = ScrapeTotalspace{}
	// 全局盘匣raid0总空间,即裸容量
	MagazinesTotalSpaceRaid0 = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_magazines_total_space_raid0"),
		"全部盘匣以raid0级别计算总空间，即裸容量，单位:bytes",
		[]string{}, nil,
	)
	// 全局盘匣实际总空间
	MagazinesTotalSpace = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_magazines_total_space"),
		"全部盘匣实际总空间,即做了raid和没做raid的所有盘匣的总容量,单位:Byte",
		[]string{}, nil,
	)
	// 全局盘匣实际剩余总空间
	MagazinesTotalAvailableSpace = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_magazines_total_available_space"),
		"全部盘匣实际可用空间,单位:Byte",
		[]string{}, nil,
	)
	// 全部盘匣槽位数
	MagzainesTotalSlotCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_magazines_total_slot_count"),
		"全部盘匣槽位数",
		[]string{}, nil,
	)
	// 全局盘匣总数
	MagazinesTotalCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_magazines_total_count"),
		"全部盘匣总数",
		[]string{}, nil,
	)
	// 全局盘匣已使用数
	MagazinesUsedCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_magazines_used_count"),
		"全部盘匣中,已使用的总数",
		[]string{}, nil,
	)
	// 全部盘匣中，未使用的总数
	MagazinesFreeCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_magazines_free_count"),
		"全部盘匣中,未使用的总数",
		[]string{}, nil,
	)
	// 全局盘匣异常数
	MagazinesExceptionCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_magazines_exception_count"),
		"全部盘匣中,异常状态的总数",
		[]string{}, nil,
	)
)

// ScrapeTotalspace is
type ScrapeTotalspace struct{}

// Name is
func (ScrapeTotalspace) Name() string {
	return "gdas_magazines_info"
}

// Help is
func (ScrapeTotalspace) Help() string {
	return "Gdas Magazines Info"
}

// Scrape is
func (ScrapeTotalspace) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	var magazinesSpace magazinesSpaceData

	url := "/api/gdas/totalspace"
	method := "POST"
	respBody, err := client.Request(method, url, nil)
	if err != nil {
		return err
	}
	//fmt.Printf("####************* %v", respBody)
	err = json.Unmarshal(respBody, &magazinesSpace)
	if err != nil {
		return err
	}
	// 全局盘匣raid0总空间,即裸容量
	ch <- prometheus.MustNewConstMetric(MagazinesTotalSpaceRaid0, prometheus.GaugeValue, float64(magazinesSpace.TotalSpaceRaid0))
	// 全局盘匣实际总空间
	ch <- prometheus.MustNewConstMetric(MagazinesTotalSpace, prometheus.GaugeValue, float64(magazinesSpace.TotalSpace))
	// 全局盘匣实际剩余总空间
	ch <- prometheus.MustNewConstMetric(MagazinesTotalAvailableSpace, prometheus.GaugeValue, float64(magazinesSpace.TotalAvailableSpace))
	// 全部盘匣槽位数
	ch <- prometheus.MustNewConstMetric(MagzainesTotalSlotCount, prometheus.GaugeValue, float64(magazinesSpace.TotalSlotCount))
	// 全局盘匣总数
	ch <- prometheus.MustNewConstMetric(MagazinesTotalCount, prometheus.GaugeValue, float64(magazinesSpace.TotalMgzCount))
	// 全局盘匣已使用数
	ch <- prometheus.MustNewConstMetric(MagazinesUsedCount, prometheus.GaugeValue, float64(magazinesSpace.UsedMgzCount))
	// 全部盘匣未使用的总数
	ch <- prometheus.MustNewConstMetric(MagazinesFreeCount, prometheus.GaugeValue, float64(magazinesSpace.FreeMgzCount))
	// 全局盘匣异常数
	ch <- prometheus.MustNewConstMetric(MagazinesExceptionCount, prometheus.GaugeValue, float64(magazinesSpace.ExceptionMgzCount))
	return nil
}

type magazinesSpaceData struct {
	// TotalSpaceRaid0 全部盘匣以raid0级别计算总空间，即裸容量，单位:bytes
	TotalSpaceRaid0 int64 `json:"totalSpaceRaid0"`
	// TotalSlotCount 全部槽位数，即整个集群最多可以插入多少块盘匣
	TotalSlotCount int `json:"totalSlotCount"`
	// TotalSpace 全部盘匣实际总空间,即做了raid和没做raid的所有盘匣的总容量,单位:Byte
	TotalSpace int64 `json:"totalSpace"`
	// Result 接口返回结果
	Result string `json:"result"`
	// ExceptionMgzCount 全部盘匣中,异常状态的总数
	ExceptionMgzCount int `json:"exceptionMgzCount"`
	// TotalAvailableSpace 全部盘匣实际可用空间,单位:Byte
	TotalAvailableSpace int64 `json:"totalAvailableSpace"`
	// FreeMgzCount 全部盘匣中,未使用的总数
	FreeMgzCount int `json:"freeMgzCount"`
	// TotalMgzCount 全部盘匣总数
	TotalMgzCount int `json:"totalMgzCount"`
	// UsedMgzCount 全部盘匣中,已使用的总数
	UsedMgzCount int `json:"usedMgzCount"`
}
