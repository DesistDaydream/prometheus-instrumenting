package collector

import (
	"encoding/json"
	"strconv"

	"github.com/DesistDaydream/prometheus-instrumenting/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// check interface
	_ scraper.CommonScraper = ScrapeDas{}
	// 全局盘库总数
	DasTotalCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_das_total_count"),
		"全局盘库总数",
		[]string{}, nil,
	)
	// 盘库状态
	DasStatus = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_das_status"),
		"盘库状态.0-连接正常,-203-盘匣弹出中,-210-仓架解锁中,-202-系统繁忙,-102-断开连接,-100&&-103-识别中",
		[]string{"dam_name", "ip", "da_name", "da_no", "da_vendor"}, nil,
	)
	// 盘库注册、断开状态
	DasOffline = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_das_offline"),
		"盘库注册、断开状态.0-已断开,1-已注册",
		[]string{"dam_name", "ip", "da_name", "da_no", "da_vendor"}, nil,
	)
	// 盘库所具有的槽位总数
	DasSlotCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_das_slot_count"),
		"盘库所具有的槽位总数，槽位就是可以插盘匣的位置",
		[]string{"dam_name", "ip", "da_name", "da_no", "da_vendor"}, nil,
	)
)

// ScrapeDas 是将要实现 Scraper 接口的一个 Metric 结构体
type ScrapeDas struct{}

// Name 指定自己定义的 抓取器 的名字，与 Metric 的名字不是一个概念，但是一般保持一致
// 该方法用于为 ScrapeDas 结构体实现 Scraper 接口
func (ScrapeDas) Name() string {
	return "gdas_das_info"
}

// Help 指定自己定义的 抓取器 的帮助信息，这里的 Help 的内容将会作为命令行标志的帮助信息。与 Metric 的 Help 不是一个概念。
// 该方法用于为 ScrapeDas 结构体实现 Scraper 接口
func (ScrapeDas) Help() string {
	return "Gdas Das Info"
}

// Scrape 从客户端采集数据，并将其作为 Metric 通过 channel(通道) 发送。主要就是采集 Gdas 集群信息的具体行为。
// 该方法用于为 ScrapeDas 结构体实现 Scraper 接口
func (ScrapeDas) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	// 声明需要绑定的 响应体 与 结构体
	var dasdata dasData

	// 根据 URI 获取 Response Body
	url := "/api/gdas/das"
	method := "POST"
	respBody, err := client.Request(method, url, nil)
	if err != nil {
		return err
	}
	// 绑定 Body 与 struct
	err = json.Unmarshal(respBody, &dasdata)
	if err != nil {
		return err
	}
	//fmt.Printf("***打印与body绑定的data结构体 %v\n", dasdata)

	// 全局盘库总数
	ch <- prometheus.MustNewConstMetric(DasTotalCount, prometheus.GaugeValue, float64(len(dasdata.DaInfo)))

	for i := 0; i < len(dasdata.DaInfo); i++ {
		//盘库状态
		ch <- prometheus.MustNewConstMetric(DasStatus, prometheus.GaugeValue, float64(dasdata.DaInfo[i].DaStatus),
			dasdata.DaInfo[i].DamName,
			dasdata.DaInfo[i].IP,
			dasdata.DaInfo[i].DaName,
			strconv.Itoa(dasdata.DaInfo[i].DaNo),
			dasdata.DaInfo[i].DaVendor,
		)
		// 盘库注册、断开状态
		ch <- prometheus.MustNewConstMetric(DasOffline, prometheus.GaugeValue, float64(dasdata.DaInfo[i].Offline),
			dasdata.DaInfo[i].DamName,
			dasdata.DaInfo[i].IP,
			dasdata.DaInfo[i].DaName,
			strconv.Itoa(dasdata.DaInfo[i].DaNo),
			dasdata.DaInfo[i].DaVendor,
		)
		// 盘库所具有的槽位总数
		ch <- prometheus.MustNewConstMetric(DasSlotCount, prometheus.GaugeValue, float64(dasdata.DaInfo[i].SlotCount),
			dasdata.DaInfo[i].DamName,
			dasdata.DaInfo[i].IP,
			dasdata.DaInfo[i].DaName,
			strconv.Itoa(dasdata.DaInfo[i].DaNo),
			dasdata.DaInfo[i].DaVendor,
		)
	}

	return nil
}

type dasData struct {
	Result string   `json:"result"`
	DaInfo []daInfo `json:"daInfo"`
}
type daInfo struct {
	// DamName 盘库所在节点名称
	DamName string `json:"damName"`
	// ChangerSerialNumber ???盘库序列号????这不是机械手的信息么???
	ChangerSerialNumber string `json:"changerSerialNumber"`
	// IP 盘库所在节点 IP
	IP string `json:"ip"`
	// DaName 盘库型号
	DaName string `json:"daName"`
	// DaStatus 盘库状态。0 正常，-203 盘匣弹出中，-210 仓架解锁中，-202 系统繁忙，-102 断开连接，-100和-103 识别中
	DaStatus int `json:"daStatus"`
	// DaNo ！！！未知！！！好像每个节点第一个盘库的号就是0，所以看到的就都是0
	DaNo int `json:"daNo"`
	// DaVendor 厂商信息
	DaVendor string `json:"daVendor"`
	// Offline 盘库注册、断开状态。0 断开，1 已注册
	Offline int `json:"offline"`
	// SlotCount 槽位总数
	SlotCount int `json:"slotCount"`
}
