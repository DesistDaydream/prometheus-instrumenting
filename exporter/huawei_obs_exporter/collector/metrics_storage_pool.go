package collector

import (
	"encoding/json"
	"strconv"

	"github.com/DesistDaydream/prometheus-instrumenting/exporter/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	_ scraper.CommonScraper = ScrapeStoragePool{}

	storagePoolStatus = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "storage_pool_status"),
		"存储池状态,0：正常,1：故障,2：写保护,3：停止,4：故障且写保护,5：数据迁移,7：降级,8：数据重构",
		[]string{"pool_id"}, nil,
	)

	storagePoolTotalCapacity = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "storage_pool_total_capacity"),
		"存储池总容量,MiB",
		[]string{"pool_id"}, nil,
	)
	storagePoolUsedCapacity = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "storage_pool_used_capacity"),
		"存储池已用容量,MiB",
		[]string{"pool_id"}, nil,
	)
)

// ScrapeStoragePool 是将要实现 Scraper 接口的一个 Metric 结构体
type ScrapeStoragePool struct{}

// Name 指定自己定义的 抓取器 的名字，与 Metric 的名字不是一个概念，但是一般保持一致
func (ScrapeStoragePool) Name() string {
	return "storage_pool_info"
}

// Help 指定自己定义的 抓取器 的帮助信息，这里的 Help 的内容将会作为命令行标志的帮助信息。与 Metric 的 Help 不是一个概念。
func (ScrapeStoragePool) Help() string {
	return "HWObs Storage Pool info"
}

// Scrape 从客户端采集数据，并将其作为 Metric 通过 channel(通道) 发送。主要就是采集 HWObs 集群信息的具体行为。
func (ScrapeStoragePool) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	var (
		storagePoolRespBody []byte
		storagePoolData     storagePoolData
	)

	url := "/dsware/service/resource/queryStoragePool"
	if storagePoolRespBody, err = client.Request("GET", url, nil); err != nil {
		return err
	}

	if err = json.Unmarshal(storagePoolRespBody, &storagePoolData); err != nil {
		return err
	}

	logrus.Debugf("存储池总数：%v", len(storagePoolData.StoragePools))

	for _, storagePool := range storagePoolData.StoragePools {
		// 存储池状态
		ch <- prometheus.MustNewConstMetric(storagePoolStatus, prometheus.GaugeValue, float64(storagePool.PoolStatus),
			strconv.Itoa(storagePool.PoolID),
		)
		// 存储池总容量
		ch <- prometheus.MustNewConstMetric(storagePoolTotalCapacity, prometheus.GaugeValue, float64(storagePool.TotalCapacity),
			strconv.Itoa(storagePool.PoolID),
		)
		// 存储池已用容量
		ch <- prometheus.MustNewConstMetric(storagePoolUsedCapacity, prometheus.GaugeValue, float64(storagePool.UsedCapacity),
			strconv.Itoa(storagePool.PoolID),
		)
	}
	return nil
}

// storagePoolData 存储 HWObs Cluster 相关信息的 Response Body 的数据
type storagePoolData struct {
	Result       int            `json:"result"`
	StoragePools []storagePools `json:"storagePools"`
}
type storagePools struct {
	ServiceType                       int         `json:"serviceType"`
	ReplicationFactor                 interface{} `json:"replicationFactor"`
	CompressionSaved                  int         `json:"compressionSaved"`
	PoolTier                          int         `json:"poolTier"`
	ProtectSwitch                     int         `json:"protectSwitch"`
	FreeCapacityRate                  float64     `json:"freeCapacityRate"`
	RedundancyPolicy                  string      `json:"redundancyPolicy"`
	PoolMode                          int         `json:"poolMode"`
	UsedCapacityAfterDedup            int         `json:"usedCapacityAfterDedup"`
	CellSize                          int         `json:"cellSize"`
	SecurityLevel                     string      `json:"securityLevel"`
	TotalCapacity                     int64       `json:"totalCapacity"`
	EcCacheRate                       interface{} `json:"ecCacheRate"`
	DataReductionRatio                float64     `json:"dataReductionRatio"`
	CacheMediaType                    string      `json:"cacheMediaType"`
	StoragePoolID                     int         `json:"storagePoolId"`
	PoolName                          string      `json:"poolName"`
	AllocatedCapacity                 int         `json:"allocatedCapacity"`
	DeduplicationSaved                int         `json:"deduplicationSaved"`
	UsedCapacity                      int64       `json:"usedCapacity"`
	NumDataUnits                      int         `json:"numDataUnits"`
	NumFaultTolerance                 int         `json:"numFaultTolerance"`
	MaxProtectFault                   int         `json:"maxProtectFault"`
	SupportEncryptForMainStorageMedia int         `json:"supportEncryptForMainStorageMedia"`
	EdsServiceStatus                  int         `json:"edsServiceStatus"`
	MediaType                         string      `json:"mediaType"`
	PhysicalTotalCapacity             int         `json:"physicalTotalCapacity"`
	PoolType                          interface{} `json:"poolType"`
	PoolSpec                          interface{} `json:"poolSpec"`
	PoolStatus                        int         `json:"poolStatus"`
	EcCacheMediaType                  interface{} `json:"ecCacheMediaType"`
	MarkDelCapacity                   int         `json:"markDelCapacity"`
	ReductionInvolvedCapacity         int         `json:"reductionInvolvedCapacity"`
	CompressionRatio                  float64     `json:"compressionRatio"`
	StorageCacheRate                  interface{} `json:"storageCacheRate"`
	UsedCapacityRate                  float64     `json:"usedCapacityRate"`
	DeduplicationRatio                float64     `json:"deduplicationRatio"`
	CompressionAlgorithm              string      `json:"compressionAlgorithm"`
	PoolID                            int         `json:"poolId"`
	Progress                          int         `json:"progress"`
	WritableCapacity                  int         `json:"writableCapacity"`
	NumParityUnits                    int         `json:"numParityUnits"`
	EncryptType                       int         `json:"encryptType"`
	DiskPoolTierInfo                  string      `json:"diskPoolTierInfo"`
}
