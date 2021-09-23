package collector

import (
	"encoding/json"
	"strconv"

	"github.com/DesistDaydream/exporter/simulate_mysql_exporter/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	_ scraper.CommonScraper = ScrapeDisk{}

	diskCount = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "disk_count"),
		"集群中磁盘总数",
		[]string{}, nil,
	)

	diskStatus = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "disk_status"),
		"集群中磁盘状态",
		[]string{"disk_role", "disk_slot", "disk_type", "node_ip"}, nil,
	)
)

// ScrapeDisk 是将要实现 Scraper 接口的一个 Metric 结构体
type ScrapeDisk struct{}

// Name 指定自己定义的 抓取器 的名字，与 Metric 的名字不是一个概念，但是一般保持一致
func (ScrapeDisk) Name() string {
	return "disk_info"
}

// Help 指定自己定义的 抓取器 的帮助信息，这里的 Help 的内容将会作为命令行标志的帮助信息。与 Metric 的 Help 不是一个概念。
func (ScrapeDisk) Help() string {
	return "HWObs Cluster Disk info"
}

// Scrape 从客户端采集数据，并将其作为 Metric 通过 channel(通道) 发送。主要就是采集 HWObs 集群信息的具体行为。
func (ScrapeDisk) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	var (
		// node 信息
		nodeInfoRespBody []byte
		nodeInfoData     nodeInfoData
		nodeIPList       []string
		// disk 信息
		diskInfoRespBody []byte
		diskInfoData     diskInfoData
		diskCountSum     int
	)

	// 获取节点的 IP 列表
	nodeIPUrl := "/dsware/service/getNodeInfoForHealthCheckTool"
	if nodeInfoRespBody, err = client.Request("GET", nodeIPUrl, nil); err != nil {
		return err
	}
	if err = json.Unmarshal(nodeInfoRespBody, &nodeInfoData); err != nil {
		return err
	}
	for _, nodeInfo := range nodeInfoData.NodeInfo {
		nodeIPList = append(nodeIPList, nodeInfo.NodeIP)
	}
	logrus.Debugf("所有节点 IP 列表：%v", nodeIPList)

	// 根据 NodeIPs 来逐一获取每个节点上的信息
	for _, nodeIP := range nodeIPList {
		diskInfoUrl := "/dsware/service/resource/queryDiskInfo?ip=" + nodeIP
		if diskInfoRespBody, err = client.Request("GET", diskInfoUrl, nil); err != nil {
			return err
		}
		if err = json.Unmarshal(diskInfoRespBody, &diskInfoData); err != nil {
			return err
		}

		diskCountSum = diskCountSum + len(diskInfoData.Disks)
		// HWObs集群中磁盘状态
		for _, disk := range diskInfoData.Disks {
			ch <- prometheus.MustNewConstMetric(diskStatus, prometheus.GaugeValue, float64(disk.DiskStatus),
				disk.DiskRole,
				strconv.Itoa(disk.DiskSlot),
				disk.DiskType,
				nodeIP,
			)
		}

	}

	// url := "/dsware/service/resource/queryAllDisk"
	// if respBody, err = client.Request("GET", url, nil); err != nil {
	// 	return err
	// }

	// if err = json.Unmarshal(respBody, &diskData); err != nil {
	// 	return err
	// }

	// for ip, diskInfos := range diskData.Disks {
	// 	// 计算磁盘总数
	// 	diskCountSum = diskCountSum + len(diskInfos)
	// 	fmt.Println(ip)
	// }

	// HWObs集群中磁盘总数
	ch <- prometheus.MustNewConstMetric(diskCount, prometheus.GaugeValue, float64(diskCountSum))

	return nil
}

// 集群节点数据
type nodeInfoData struct {
	Result   int        `json:"result"`
	NodeInfo []NodeInfo `json:"NodeInfo"`
}
type NodeInfo struct {
	NodeName string `json:"NodeName"`
	NodeType int    `json:"NodeType"`
	NodeIP   string `json:"NodeIP"`
}

// 磁盘信息数据
type diskInfoData struct {
	Result    int               `json:"result"`
	Disks     []disks           `json:"disks"`
	Pools     diskInfoDataPools `json:"pools"`
	PoolError poolError         `json:"poolError"`
	PoolName  string            `json:"poolName"`
}
type diskInfoDataPools struct {
	PoolID int `json:"poolId"`
}

type disks struct {
	DiskExist            int           `json:"diskExist"`
	DiskSn               string        `json:"diskSn"`
	DiskSlot             int           `json:"diskSlot"`
	DiskStatus           int           `json:"diskStatus"`
	DiskType             string        `json:"diskType"`
	DiskSize             int           `json:"diskSize"`
	DiskUsedSize         int           `json:"diskUsedSize"`
	MediaCapacityForByte int64         `json:"mediaCapacityForByte"`
	DiskRole             string        `json:"diskRole"`
	SupportEncrypt       int           `json:"supportEncrypt"`
	SsdDetail            []interface{} `json:"ssdDetail"`
	Pools                []pools       `json:"pools"`
	DiskModel            string        `json:"diskModel"`
	Version              string        `json:"version"`
}
type pools struct {
	PoolID   int `json:"poolId"`
	PoolSize int `json:"poolSize"`
	Status   int `json:"status"`
}
type poolError struct {
}
