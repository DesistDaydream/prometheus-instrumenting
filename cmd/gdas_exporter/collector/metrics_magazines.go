package collector

import (
	"encoding/json"
	"strconv"

	"github.com/DesistDaydream/prometheus-instrumenting/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// check interface
	_ scraper.CommonScraper = ScrapeMagazines{}

	// 盘匣状态
	MagazinesStatus = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "magazines_status"),
		"盘匣状态.0-正常,1-异常",
		[]string{"dam_name", "ip", "da_name", "da_no", "rfid", "slot_no", "pool_name"}, nil,
	)
	// 盘匣是否已满
	MagazinesFull = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "magazines_full"),
		"盘匣空间是否已满.0-未满,1-已满",
		[]string{"dam_name", "ip", "da_name", "da_no", "rfid", "slot_no", "pool_name"}, nil,
	)
	// 盘匣是否已被分配
	MagazinesRfidSts = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "magazines_rfid_sts"),
		"盘匣是否已被分配.0-未分配,1-已分配",
		[]string{"dam_name", "ip", "da_name", "da_no", "rfid", "slot_no", "pool_name"}, nil,
	)
)

// ScrapeMagazines 是将要实现 Scraper 接口的一个 Metric 结构体
type ScrapeMagazines struct{}

// Name 指定自己定义的 抓取器 的名字，与 Metric 的名字不是一个概念，但是一般保持一致
// 该方法用于为 ScrapeMagazines 结构体实现 Scraper 接口
func (ScrapeMagazines) Name() string {
	return "gdas_magazines_info"
}

// Help 指定自己定义的 抓取器 的帮助信息，这里的 Help 的内容将会作为命令行标志的帮助信息。与 Metric 的 Help 不是一个概念。
// 该方法用于为 ScrapeMagazines 结构体实现 Scraper 接口
func (ScrapeMagazines) Help() string {
	return "Gdas Magazines Info"
}

// Scrape 从客户端采集数据，并将其作为 Metric 通过 channel(通道) 发送。主要就是采集 Gdas 集群信息的具体行为。
// 该方法用于为 ScrapeMagazines 结构体实现 Scraper 接口
func (ScrapeMagazines) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	// 声明需要绑定的 响应体 与 结构体
	var (
		respBody  []byte
		magazines Magazines
	)

	// 根据 URI 获取 Response Body
	url := "/v1/magazines"
	if respBody, err = client.Request("GET", url, nil); err != nil {
		return err
	}

	// 绑定 Body 与 struct
	if err = json.Unmarshal(respBody, &magazines); err != nil {
		return err
	}

	for i := 0; i < len(magazines.Rfid); i++ {
		// 盘匣状态
		ch <- prometheus.MustNewConstMetric(MagazinesStatus, prometheus.GaugeValue, float64(magazines.Rfid[i].Status),
			magazines.Rfid[i].DamName,
			magazines.Rfid[i].ServerIP,
			magazines.Rfid[i].DaName,
			strconv.Itoa(magazines.Rfid[i].DaNo),
			magazines.Rfid[i].Rfid,
			strconv.Itoa(magazines.Rfid[i].SlotNo),
			magazines.Rfid[i].PoolName,
		)
		// 盘匣空间是否已满
		ch <- prometheus.MustNewConstMetric(MagazinesFull, prometheus.GaugeValue, float64(magazines.Rfid[i].Full),
			magazines.Rfid[i].DamName,
			magazines.Rfid[i].ServerIP,
			magazines.Rfid[i].DaName,
			strconv.Itoa(magazines.Rfid[i].DaNo),
			magazines.Rfid[i].Rfid,
			strconv.Itoa(magazines.Rfid[i].SlotNo),
			magazines.Rfid[i].PoolName,
		)
		// 盘匣是否已分配
		ch <- prometheus.MustNewConstMetric(MagazinesRfidSts, prometheus.GaugeValue, float64(magazines.Rfid[i].RfidSts),
			magazines.Rfid[i].DamName,
			magazines.Rfid[i].ServerIP,
			magazines.Rfid[i].DaName,
			strconv.Itoa(magazines.Rfid[i].DaNo),
			magazines.Rfid[i].Rfid,
			strconv.Itoa(magazines.Rfid[i].SlotNo),
			magazines.Rfid[i].PoolName,
		)
	}
	return nil
}

// Magazines is
type Magazines struct {
	Result string `json:"result"`
	Rfid   []Rfid `json:"rfid"`
}

// Rfid is
type Rfid struct {
	Rfid     string   `json:"rfid"`
	Barcode  string   `json:"barcode"`
	DaName   string   `json:"daName"`
	PoolName string   `json:"poolName"`
	Full     int      `json:"full"`
	Format   int      `json:"format"`
	RfidSts  int      `json:"rfidSts"`
	Status   int      `json:"status"`
	SlotNo   int      `json:"slotNo"`
	DaNo     int      `json:"daNo"`
	CpGroup  []string `json:"cpGroup"`
	Offline  int      `json:"offline"`
	ServerIP string   `json:"serverIp"`
	DamName  string   `json:"damName"`
}
