package collector

import (
	"encoding/json"
	"strconv"

	"github.com/DesistDaydream/prometheus-instrumenting/pkg/scraper"
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
		[]string{"dam_name", "ip"}, nil,
	)
	// 节点总缓存容量
	NodeTotalCacheSize = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_node_total_cache_size"),
		"节点总缓存容量,单位:Byte",
		[]string{"dam_name", "ip"}, nil,
	)
	// 节点已用缓存容量
	NodeUsedCacheSize = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_node_used_cache_size"),
		"节点已用缓存容量,单位:Byte",
		[]string{"dam_name", "ip"}, nil,
	)
	// 节点未用缓存容量
	NodeUnusedCacheSize = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_node_unused_cache_size"),
		"节点未用缓存容量,单位:Byte",
		[]string{"dam_name", "ip"}, nil,
	)
	// 盘库中盘库中已用盘匣数量
	magazineUsedCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_das_magazine_used_count"),
		"盘库中已用盘匣数量",
		[]string{"dam_name", "ip", "da_name", "da_no"}, nil,
	)

	// 盘库中未用盘匣数量
	magazineFreeCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_das_magazine_free_count"),
		"盘库中未用盘匣数量",
		[]string{"dam_name", "ip", "da_name", "da_no"}, nil,
	)
	// 盘库中异常盘匣数量
	magazineExcpCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_das_magazine_excp_count"),
		"盘库中异常盘匣数量",
		[]string{"dam_name", "ip", "da_name", "da_no"}, nil,
	)
	// 盘库中机械手状态
	changerStatus = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_das_changer_status"),
		"盘库中机械手的状态:0-寿命良好,1-寿命警告,2-寿命已到",
		[]string{"dam_name", "ip", "da_name", "da_no", "changer_serial"}, nil,
	)
	// 盘库中机械手状态
	changerUsedPercent = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_das_changer_used_percent"),
		"盘库中机械手使用百分比，该值需要除以 100",
		[]string{"dam_name", "ip", "da_name", "da_no", "changer_serial"}, nil,
	)
	// TODO:盘库中机械手数量

	// 光驱状态
	driveStatus = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_das_drive_status"),
		"盘库中光驱的状态:0-寿命良好,1-寿命警告,2-寿命已到",
		[]string{"dam_name", "ip", "da_name", "da_no", "drive_serial"}, nil,
	)
	// 光驱状态
	driveUsedPercent = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_das_drive_used_percent"),
		"盘库中光驱使用百分比，该值需要除以 100",
		[]string{"dam_name", "ip", "da_name", "da_no", "drive_serial"}, nil,
	)
	// TODO:盘库中光驱数量
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
	// ####### 获取分布式节点信息,即node概况 #######
	// #############################################
	var (
		nodedata    nodeData
		cacheData   cacheData
		nodeDasData nodeDasData
		nodeIPList  []string
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

	// ################################################
	// # 通过分布式节点信息的内容，逐一获取每个节点的指标 #
	// ################################################
	for index, nodeIP := range nodeIPList {
		// 获取节点的状态
		ch <- prometheus.MustNewConstMetric(NodeStatus, prometheus.GaugeValue, float64(nodedata.NodeList[index].Status),
			nodedata.NodeList[index].DamName,
			nodeIP,
		)

		// ################################################
		// ####### 循环每个节点，逐一获取节点的缓存数据 ######
		// ################################################
		cacheUrl := "/api/gdas/cache/node/" + nodeIP
		cacheMethod := "POST"
		respBodyCacheNode, err := client.Request(cacheMethod, cacheUrl, nil)
		if err != nil {
			return err
		}
		err = json.Unmarshal(respBodyCacheNode, &cacheData)
		if err != nil {
			return err
		}

		//节点总缓存容量
		ch <- prometheus.MustNewConstMetric(NodeTotalCacheSize, prometheus.GaugeValue, float64(cacheData.TotalCacheSize),
			nodedata.NodeList[index].DamName,
			nodeIP,
		)
		//节点已用缓存容量
		ch <- prometheus.MustNewConstMetric(NodeUsedCacheSize, prometheus.GaugeValue, float64(cacheData.UsedCacheSize),
			nodedata.NodeList[index].DamName,
			nodeIP,
		)
		//节点未用缓存容量
		ch <- prometheus.MustNewConstMetric(NodeUnusedCacheSize, prometheus.GaugeValue, float64(cacheData.UnUsedCacheSize),
			nodedata.NodeList[index].DamName,
			nodeIP,
		)

		// ################################################
		// #### 循环每个节点，逐一获取节点下每个盘库的信息 ####
		// ################################################
		dasUrl := "/api/gdas/das/node/" + nodeIP
		dasMethod := "POST"
		respBodyDasNode, err := client.Request(dasMethod, dasUrl, nil)
		if err != nil {
			return err
		}
		err = json.Unmarshal(respBodyDasNode, &nodeDasData)
		if err != nil {
			return err
		}

		// 每个节点下有多个盘库，所以循环每个盘库以获取指标
		for j := 0; j < len(nodeDasData.DaList); j++ {
			// 盘库中已用盘匣数量
			ch <- prometheus.MustNewConstMetric(magazineUsedCount, prometheus.GaugeValue, float64(nodeDasData.DaList[j].MagazineUsedCount),
				nodedata.NodeList[index].DamName,
				nodeIP,
				nodeDasData.DaList[j].Name,
				strconv.Itoa(nodeDasData.DaList[j].DaNo),
			)
			// 盘库中未用盘匣数量
			ch <- prometheus.MustNewConstMetric(magazineFreeCount, prometheus.GaugeValue, float64(nodeDasData.DaList[j].MagazineFreeCount),
				nodedata.NodeList[index].DamName,
				nodeIP,
				nodeDasData.DaList[j].Name,
				strconv.Itoa(nodeDasData.DaList[j].DaNo),
			)
			// 盘库中异常盘匣数量
			ch <- prometheus.MustNewConstMetric(magazineExcpCount, prometheus.GaugeValue, float64(nodeDasData.DaList[j].MagazineExcpCount),
				nodedata.NodeList[index].DamName,
				nodeIP,
				nodeDasData.DaList[j].Name,
				strconv.Itoa(nodeDasData.DaList[j].DaNo),
			)

			// 循环盘库下每个机械手，以获取指标
			for k := 0; k < len(nodeDasData.DaList[j].ChangerSmartInfo); k++ {
				// 机械手状态
				ch <- prometheus.MustNewConstMetric(changerStatus, prometheus.GaugeValue, float64(nodeDasData.DaList[j].ChangerSmartInfo[k].Status),
					nodedata.NodeList[index].DamName,
					nodeIP,
					nodeDasData.DaList[j].Name,
					strconv.Itoa(nodeDasData.DaList[j].DaNo),
					nodeDasData.DaList[j].ChangerSerial,
					// strconv.Itoa(status.DaList[j].ChangerSmartInfo[k].UnitNo),
				)
				// 机械手使用百分比
				ch <- prometheus.MustNewConstMetric(changerUsedPercent, prometheus.GaugeValue, float64(nodeDasData.DaList[j].ChangerSmartInfo[k].UsedPercent),
					nodedata.NodeList[index].DamName,
					nodeIP,
					nodeDasData.DaList[j].Name,
					strconv.Itoa(nodeDasData.DaList[j].DaNo),
					nodeDasData.DaList[j].ChangerSerial,
					// strconv.Itoa(status.DaList[j].ChangerSmartInfo[k].UnitNo),
				)
			}
			// 循环盘库下每个光驱，以获取指标
			// 判断一下与光驱有关的另一个数组中的元素是否不为空
			if len(nodeDasData.DaList[j].DriveSerialList) > 0 {
				for l := 0; l < len(nodeDasData.DaList[j].DriveSmartInfo); l++ {
					// 光驱状态
					ch <- prometheus.MustNewConstMetric(driveStatus, prometheus.GaugeValue, float64(nodeDasData.DaList[j].DriveSmartInfo[l].Status),
						nodedata.NodeList[index].DamName,
						nodeIP,
						nodeDasData.DaList[j].Name,
						strconv.Itoa(nodeDasData.DaList[j].DaNo),
						nodeDasData.DaList[j].DriveSerialList[l].DriveSerial,
						// strconv.Itoa(status.DaList[j].DriveSmartInfo[l].UnitNo),
					)
					// 光驱使用百分比
					ch <- prometheus.MustNewConstMetric(driveUsedPercent, prometheus.GaugeValue, float64(nodeDasData.DaList[j].DriveSmartInfo[l].UsedPercent),
						nodedata.NodeList[index].DamName,
						nodeIP,
						nodeDasData.DaList[j].Name,
						strconv.Itoa(nodeDasData.DaList[j].DaNo),
						nodeDasData.DaList[j].DriveSerialList[l].DriveSerial,
						// strconv.Itoa(status.DaList[j].DriveSmartInfo[l].UnitNo),
					)
				}
			} else {
				logrus.Error("从 API 获取光驱指标异常,DriveSerialList 数组元素不大于0")
			}
		}
	}

	return nil
}

// /api/gdas/node/list 返回的数据
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

// /api/gdas/cache/node/{nodeIP} 返回的数据
// 每个节点的缓存信息
type cacheData struct {
	// UsedCacheSize 当前节点的已用缓存大小
	UsedCacheSize int64  `json:"usedCacheSize"`
	Result        string `json:"result"`
	// UnUsedCacheSize 当前节点的剩余缓存大小
	UnUsedCacheSize int64 `json:"unUsedCacheSize"`
	// TotalCacheSize 当前节点的总缓存大小
	TotalCacheSize int64 `json:"totalCacheSize"`
}

// /api/gdas/das/node/{nodeIP} 返回的数据
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
	// DaNo ？？？？应该节点下是盘库的序号，从0开始？？？？
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
