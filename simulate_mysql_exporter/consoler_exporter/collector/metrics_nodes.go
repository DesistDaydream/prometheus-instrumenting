package collector

import (
	"encoding/json"

	"github.com/DesistDaydream/exporter/simulate_mysql_exporter/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	_ scraper.CommonScraper = ScrapeNodes{}
	// 全局节点总数
	NodeTotalCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_node_total_count"),
		"全局节点总数",
		[]string{}, nil,
	)
	// 节点状态
	NodeStatus = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_node_status"),
		"节点状态:0-活跃,1-异常",
		[]string{"damName", "ip"}, nil,
	)
	// 节点总缓存容量
	NodeTotalCacheSize = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_node_total_cache_size"),
		"节点总缓存容量,单位:Byte",
		[]string{"ip"}, nil,
	)
	// 节点已用缓存容量
	NodeUsedCacheSize = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_node_used_cache_size"),
		"节点已用缓存容量,单位:Byte",
		[]string{"ip"}, nil,
	)
	// 节点未用缓存容量
	NodeUnusedCacheSize = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_node_unused_cache_size"),
		"节点未用缓存容量,单位:Byte",
		[]string{"ip"}, nil,
	)
	// 机械手状态
	changerStatus = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_node_changer_status"),
		"盘库中机械手的状态:0-寿命良好,1-寿命警告,2-寿命已到",
		[]string{"ip", "name", "changer_serial"}, nil,
	)
	// 机械手数量

	// 光驱状态
	driveStatus = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_node_drive_status"),
		"盘库中光驱的状态:0-寿命良好,1-寿命警告,2-寿命已到",
		[]string{"ip", "name", "drive_serial"}, nil,
	)
	// 光驱数量

)

// ScrapeNodes is
type ScrapeNodes struct{}

// Name is
func (ScrapeNodes) Name() string {
	return "gdas_nodes_info"
}

// Help is
func (ScrapeNodes) Help() string {
	return "Gdas Nodes Info"
}

// Scrape is
func (ScrapeNodes) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	// #############################################
	// ######## 获取分布式节点信息,即node概况 ########
	// #############################################
	// 获取分布式节点信息的指标
	var (
		nodedata   nodeData
		nodeIPList []string
	)
	url := "/api/gdas/node/list"
	method := "POST"
	respBodyNodeData, err := client.Request(method, url, nil)
	if err != nil {
		return err
	}
	err = json.Unmarshal(respBodyNodeData, &nodedata)
	if err != nil {
		return err
	}

	// 获取 nodeIP 列表，放到切片中
	for i := 0; i < len(nodedata.NodeList); i++ {
		nodeIPList = append(nodeIPList, nodedata.NodeList[i].IP)
	}
	logrus.Debugf("所有节点 IP 列表：%v", nodeIPList)

	// 集群中节点总数
	ch <- prometheus.MustNewConstMetric(NodeTotalCount, prometheus.GaugeValue, float64(len(nodedata.NodeList)))

	// 每个节点的状态
	for index, nodeIP := range nodeIPList {
		ch <- prometheus.MustNewConstMetric(NodeStatus, prometheus.GaugeValue, float64(nodedata.NodeList[index].Status),
			nodedata.NodeList[index].DamName,
			nodeIP,
		)
	}

	// ################################################
	// # 通过分布式节点信息的内容，逐一获取每个节点的指标 #
	// ################################################
	//
	// 循环每个节点，逐一获取节点的缓存数据
	var cacheData cacheData
	for _, nodeIP := range nodeIPList {
		url := "/api/gdas/cache/node/" + nodeIP
		method := "POST"
		respBodyCacheNode, err := client.Request(method, url, nil)
		if err != nil {
			return err
		}
		err = json.Unmarshal(respBodyCacheNode, &cacheData)
		if err != nil {
			return err
		}

		//节点总缓存容量
		ch <- prometheus.MustNewConstMetric(NodeTotalCacheSize, prometheus.GaugeValue, float64(cacheData.TotalCacheSize),
			nodeIP,
		)
		//节点已用缓存容量
		ch <- prometheus.MustNewConstMetric(NodeUsedCacheSize, prometheus.GaugeValue, float64(cacheData.UsedCacheSize),
			nodeIP,
		)
		//节点未用缓存容量
		ch <- prometheus.MustNewConstMetric(NodeUnusedCacheSize, prometheus.GaugeValue, float64(cacheData.UnUsedCacheSize),
			nodeIP,
		)
	}

	// 循环每个节点，逐一获取节点下每个盘库的信息
	var nodeDasData nodeDasData
	for _, nodeIP := range nodeIPList {
		url := "/api/gdas/das/node/" + nodeIP
		method := "POST"
		respBodyDasNode, err := client.Request(method, url, nil)
		if err != nil {
			return err
		}
		err = json.Unmarshal(respBodyDasNode, &nodeDasData)
		if err != nil {
			return err
		}

		// 循环每个盘库
		for j := 0; j < len(nodeDasData.DaList); j++ {
			for k := 0; k < len(nodeDasData.DaList[j].ChangerSmartInfo); k++ {
				//机械手状态
				ch <- prometheus.MustNewConstMetric(changerStatus, prometheus.GaugeValue, float64(nodeDasData.DaList[j].ChangerSmartInfo[k].Status),
					nodeIP,
					nodeDasData.DaList[j].Name,
					nodeDasData.DaList[j].ChangerSerial,
					// strconv.Itoa(status.DaList[j].ChangerSmartInfo[k].UnitNo),
				)
			}

			for l := 0; l < len(nodeDasData.DaList[j].DriveSmartInfo); l++ {
				//光驱状态
				ch <- prometheus.MustNewConstMetric(driveStatus, prometheus.GaugeValue, float64(nodeDasData.DaList[j].DriveSmartInfo[l].Status),
					nodeIP,
					nodeDasData.DaList[j].Name,
					nodeDasData.DaList[j].DriveSerialList[l].DriveSerial,
					// strconv.Itoa(status.DaList[j].DriveSmartInfo[l].UnitNo),
				)
			}
		}
	}
	return nil
}

// 全部节点的信息
type nodeData struct {
	Result   string     `json:"result"`
	NodeList []nodeList `json:"nodeList"`
}

// 每个节点的信息
type nodeList struct {
	// DamName 节点名称
	DamName string `json:"damName"`
	// Status 节点状态.0-活跃,1-异常
	Status int `json:"status"`
	// IP 节点IP地址
	IP string `json:"ip"`
}

// 每个节点的缓存信息
type cacheData struct {
	// UsedCacheSize 当前节点的已用缓存大小
	UsedCacheSize int64 `json:"usedCacheSize"`
	// Result 略
	Result string `json:"result"`
	// UnUsedCacheSize 当前节点的剩余缓存大小
	UnUsedCacheSize int64 `json:"unUsedCacheSize"`
	// TotalCacheSize 当前节点的总缓存大小
	TotalCacheSize int64 `json:"totalCacheSize"`
}

// 每个节点下的全部盘库信息
type nodeDasData struct {
	Result string `json:"result"`
	// DaList 节点列表
	DaList []daList `json:"daList"`
}

// 节点下每个盘库的信息
type daList struct {
	// SlotNum 盘库的槽位数
	SlotNum int `json:"slot_num"`
	// IP 盘库所在节点的IP地址
	IP string `json:"ip"`
	// MagazineUsedCount 盘库已使用盘匣个数
	MagazineUsedCount int `json:"magazineUsedCount"`
	// MagazineFreeCount 盘库未使用盘匣个数
	MagazineFreeCount int `json:"magazineFreeCount"`
	// DaStatus 盘库状态.0-连接正常,-203-盘匣弹出中,-210-仓架解锁中,-202-系统繁忙,-102-断开连接,-100&&-103-识别中
	DaStatus int `json:"daStatus"`
	// ChangerNum 盘库机械手数量
	ChangerNum int `json:"changer_num"`
	// DaNo ！！没用！！值永远为0
	DaNo int `json:"da_no"`
	// ChangerSerial 机械手序列号
	ChangerSerial string `json:"changerSerial"`
	// DriveSerialList 光驱序列号信息列表
	DriveSerialList []driveSerialList `json:"driveSerialList"`
	// Name 盘库型号
	Name string `json:"name"`
	// MagazineExcpCount 当前盘库异常盘匣个数
	MagazineExcpCount int `json:"magazineExcpCount"`
	// DriveNum 盘库光驱数量
	DriveNum int `json:"drive_num"`
	// ChangerSmartInfo 机械手的smart信息(编号、状态、使用百分比)
	ChangerSmartInfo []changerSmartInfo `json:"changerSmartInfo"`
	// DriveSmartInfo 光驱的smart信息(编号、状态、使用百分比)
	DriveSmartInfo []driveSmartInfo `json:"driveSmartInfo"`
}

// 盘库中光驱的序列号信息详情
type driveSerialList struct {
	// DriveNo 光驱号
	DriveNo int `json:"driveNo"`
	// DriveSerial 光驱序列号
	DriveSerial string `json:"driveSerial"`
}

// 盘库中每个机械手的信息
type changerSmartInfo struct {
	// UnitNo 机械手号，默认为0
	UnitNo int `json:"unitNo"`
	// UsedPercent 机械手使用百分比
	UsedPercent int `json:"usedPercent"`
	// Status 机械手状态
	Status int `json:"status"`
}

// 盘库中每个光驱的信息
type driveSmartInfo struct {
	// UnitNo 光驱号，默认为0
	UnitNo int `json:"unitNo"`
	// UsedPercent 光驱使用百分比
	UsedPercent int `json:"usedPercent"`
	// Status 光驱状态
	Status int `json:"status"`
}
