package collector

import (
	"encoding/json"
	"strconv"

	"github.com/DesistDaydream/prometheus-instrumenting/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ scraper.CommonScraper = ScrapeMagazinesMetrics{}
	// 盘匣状态
	MagazinesStatus = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_magazines_status"),
		"盘匣状态.0-正常,1-异常",
		[]string{"dam_name", "ip", "da_name", "da_no", "rfid", "slot_no", "pool_name"}, nil,
	)
	// 盘匣是否已满
	MagazinesFull = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_magazines_full"),
		"盘匣空间是否已满.0-未满,1-已满",
		[]string{"dam_name", "ip", "da_name", "da_no", "rfid", "slot_no", "pool_name"}, nil,
	)
	// 盘匣是否已被分配
	MagazinesRFIDSts = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_magazines_rfid_sts"),
		"盘匣是否已被分配.0-未分配,1-已分配",
		[]string{"dam_name", "ip", "da_name", "da_no", "rfid", "slot_no", "pool_name"}, nil,
	)
)

// ScrapeMagazinesMetrics is
type ScrapeMagazinesMetrics struct{}

// Name is
func (ScrapeMagazinesMetrics) Name() string {
	return "gdas_magazines_status"
}

// Help is
func (ScrapeMagazinesMetrics) Help() string {
	return "Gdas Magazines Status.0-normal,3-copying,4-non-system,9-anormal"
}

// Scrape is
func (ScrapeMagazinesMetrics) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	var magazines magazinesData

	url := "/api/gdas/magazine/list"
	method := "POST"
	respBody, err := client.Request(method, url, nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &magazines)
	if err != nil {
		return err
	}

	for i := 0; i < len(magazines.RFID); i++ {
		// 盘匣状态
		ch <- prometheus.MustNewConstMetric(MagazinesStatus, prometheus.GaugeValue, float64(magazines.RFID[i].Status),
			magazines.RFID[i].DamName,
			magazines.RFID[i].ServerIP,
			magazines.RFID[i].DaName,
			strconv.Itoa(magazines.RFID[i].DaNo),
			magazines.RFID[i].RFID,
			strconv.Itoa(magazines.RFID[i].SlotNo),
			magazines.RFID[i].PoolName,
		)
		// 盘匣空间是否已满
		ch <- prometheus.MustNewConstMetric(MagazinesFull, prometheus.GaugeValue, float64(magazines.RFID[i].Full),
			magazines.RFID[i].DamName,
			magazines.RFID[i].ServerIP,
			magazines.RFID[i].DaName,
			strconv.Itoa(magazines.RFID[i].DaNo),
			magazines.RFID[i].RFID,
			strconv.Itoa(magazines.RFID[i].SlotNo),
			magazines.RFID[i].PoolName,
		)
		// 盘匣是否已分配
		ch <- prometheus.MustNewConstMetric(MagazinesRFIDSts, prometheus.GaugeValue, float64(magazines.RFID[i].RFIDSts),
			magazines.RFID[i].DamName,
			magazines.RFID[i].ServerIP,
			magazines.RFID[i].DaName,
			strconv.Itoa(magazines.RFID[i].DaNo),
			magazines.RFID[i].RFID,
			strconv.Itoa(magazines.RFID[i].SlotNo),
			magazines.RFID[i].PoolName,
		)
	}
	return nil
}

type magazinesData struct {
	Result string `json:"result"`
	RFID   []RFID `json:"rfid"`
}

type RFID struct {
	// Barcode !!未知！！
	Barcode string `json:"barcode"`
	// CpGroup 副本盘匣列表
	CpGroup []string `json:"cpGroup"`
	// DaName 盘库型号
	DaName string `json:"daName"`
	// DaNo ？？？？应该节点下是盘库的序号，从0开始？？？？
	DaNo int `json:"daNo"`
	// DamName 所属盘库的节点名
	DamName string `json:"damName"`
	// Format 盘匣是否被格式化
	Format int `json:"format"`
	// Full 各个盘匣空间是否已满
	Full int `json:"full"`
	// Offline 盘匣是否离线
	Offline int `json:"offline"`
	// PoolName 盘匣所属盘匣组名称
	PoolName string `json:"poolName"`
	// RFID 射频标识符，即盘匣唯一标识符
	RFID string `json:"rfid"`
	// RFIDSts 盘匣是否被分配
	RFIDSts int `json:"rfidSts"`
	// ServerIP 该盘匣所属盘库的节点 IP
	ServerIP string `json:"serverIp"`
	// SlotNo 盘匣所在槽位号
	SlotNo int `json:"slotNo"`
	// Status 盘匣状态
	Status int `json:"status"`
}
