package collector

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/DesistDaydream/prometheus-instrumenting/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ scraper.CommonScraper = ScrapePool{}
	// 全局盘匣组总数
	PoolTotalCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_pool_total_count"),
		"全局盘匣组总数",
		[]string{}, nil,
	)
	// 盘匣组状态
	PoolStatus = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_pool_status"),
		"盘匣组状态.0-空闲,1-刻录",
		[]string{"pool_can_del_flag", "type", "default_mgz", "pool_name", "user", "pool_raidLvl", "auto_add_mgz"}, nil,
	)

	// 盘匣组存储总空间
	PoolTotalSpace = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_pool_total_space"),
		"盘匣组总存储空间,单位:Byte",
		[]string{"pool_name", "pool_raidLvl"}, nil,
	)
	// 盘匣组存储剩余空间
	PoolAvailableSpace = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_pool_available_space"),
		"盘匣组可用存储空间,单位:Byte",
		[]string{"pool_name", "pool_raidLvl"}, nil,
	)
	// 盘匣组中的盘匣数量
	PoolRfidCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_pool_rfid_count"),
		"盘匣组包含的盘匣数量",
		[]string{"pool_name", "pool_raidLvl"}, nil,
	)
)

// ScrapePool is
type ScrapePool struct{}

// Name is
func (ScrapePool) Name() string {
	return "gdas_pool_info"
}

// Help is
func (ScrapePool) Help() string {
	return "Gdas Pool Info"
}

// Scrape is
func (ScrapePool) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	var pooldata poolData

	url := "/api/gdas/pool/list"
	method := "POST"
	reqBody := `{"poolFlag":false,"poolName":"","poolType":""}`
	buf := bytes.NewBuffer([]byte(reqBody))
	respBody, err := client.Request(method, url, buf)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &pooldata)
	if err != nil {
		return err
	}

	// 全局盘匣组总数
	ch <- prometheus.MustNewConstMetric(PoolTotalCount, prometheus.GaugeValue, float64(pooldata.ResCount))

	for i := 0; i < len(pooldata.Pools); i++ {
		// 盘匣组状态
		ch <- prometheus.MustNewConstMetric(PoolStatus, prometheus.GaugeValue, float64(pooldata.Pools[i].PoolSts),
			pooldata.Pools[i].PoolCanDelFlag,
			pooldata.Pools[i].Type,
			strconv.FormatBool(pooldata.Pools[i].DefaultMgz),
			pooldata.Pools[i].PoolName,
			pooldata.Pools[i].User,
			strconv.Itoa(pooldata.Pools[i].PoolRaidLvl),
			strconv.FormatBool(pooldata.Pools[i].AutoAddMgz),
		)
		// 盘匣组存储总空间
		ch <- prometheus.MustNewConstMetric(PoolTotalSpace, prometheus.GaugeValue, float64(pooldata.Pools[i].PoolTotalSpace),
			pooldata.Pools[i].PoolName,
			strconv.Itoa(pooldata.Pools[i].PoolRaidLvl),
		)
		// 盘匣组存储剩余空间
		ch <- prometheus.MustNewConstMetric(PoolAvailableSpace, prometheus.GaugeValue, float64(pooldata.Pools[i].PoolAvailableSpace),
			pooldata.Pools[i].PoolName,
			strconv.Itoa(pooldata.Pools[i].PoolRaidLvl),
		)
		// 盘匣组中的盘匣数量
		ch <- prometheus.MustNewConstMetric(PoolRfidCount, prometheus.GaugeValue, float64(pooldata.Pools[i].RfidCount),
			pooldata.Pools[i].PoolName,
			strconv.Itoa(pooldata.Pools[i].PoolRaidLvl),
		)
	}
	return nil
}

type poolData struct {
	Result string `json:"result"`
	// 全局盘匣组总数
	ResCount int `json:"res_count"`
	// 盘匣组信息列表
	Pools []pools `json:"pools"`
}
type pools struct {
	// PoolCanDelFlag 盘匣组是否可删除.0-可删除,1-存在数据不可删除,2-存在桶不可删除
	PoolCanDelFlag string `json:"poolCanDel_flag"`
	// Type 盘匣组存储类型
	Type string `json:"type"`
	// DefaultMgz 是否是默认盘匣组
	DefaultMgz bool `json:"defaultMgz"`
	// PoolTotalSpace 盘匣组总空间
	PoolTotalSpace int64 `json:"pool_total_space"`
	// PoolName 盘匣组名称
	PoolName string `json:"pool_name"`
	// PoolSts 盘匣组状态.0-空闲,1-刻录
	PoolSts int `json:"pool_sts"`
	// RfidCount 盘匣组有用的盘匣数量
	RfidCount int `json:"rfid_count"`
	// PoolAvailableSpace 盘匣组可用空间
	PoolAvailableSpace int64 `json:"pool_available_space"`
	// User 盘匣组所属用户的名称
	User string `json:"user"`
	// PoolRaidLvl 盘匣组的raid级别
	PoolRaidLvl int `json:"pool_raidLvl"`
	// AutoAddMgz 是否允许自动追加盘匣
	AutoAddMgz bool `json:"autoAddMgz"`
}
