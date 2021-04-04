package collector

import (
	"encoding/json"
	"fmt"

	"github.com/DesistDaydream/exporter/simulate_mysql_exporter/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// check interface
	_ scraper.CommonScraper = ScrapeMagazines{}

	// 设置 Metric 的基本信息
	magazines = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "magazines_info"),
		"Gdas Magazines Info",
		[]string{"comments"}, nil,
	)
)

// ScrapeMagazines 是将要实现 Scraper 接口的一个 Metric 结构体
type ScrapeMagazines struct{}

// Name 指定自己定义的 抓取器 的名字，与 Metric 的名字不是一个概念，但是一般保持一致
// 该方法用于为 ScrapeMagazines 结构体实现 Scraper 接口
func (ScrapeMagazines) Name() string {
	return "magazines_info"
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
		respBody []byte
		data     Magazines
	)

	// 根据 URI 获取 Response Body
	url := "/v1/magazines"
	if respBody, err = client.Request("GET", url, nil); err != nil {
		return err
	}

	// 绑定 Body 与 struct
	if err = json.Unmarshal(respBody, &data); err != nil {
		return err
	}

	// 根据 Response Body 获取盘匣使用量
	var undistributedCount float64
	for i := 0; i < len(data.Rfid); i++ {
		rfidSts := data.Rfid[i].RfidSts
		// fmt.Println(test)
		// 如果 rfidSts 为 0，则计数器加1
		if rfidSts == 1 {
			undistributedCount = undistributedCount + 1
		}
	}

	fmt.Printf("当前分配了 %v 个盘匣\n", undistributedCount)
	ch <- prometheus.MustNewConstMetric(magazines, prometheus.GaugeValue, undistributedCount, "undistributedCount")
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
