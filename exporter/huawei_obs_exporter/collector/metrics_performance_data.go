package collector

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"

	"github.com/DesistDaydream/prometheus-instrumenting/exporter/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// 性能类型的标识符
const (
	DELETERequestPerSecond = 540
	GETRequestPerSecond    = 543
	PUTRequestPerSecond    = 546
	POSTRequestPerSecond   = 1064
	ReadBandwidth          = 50001
	WriteBandwidth         = 50002
	TotalBandwidth         = 50003
)

var (
	_ scraper.CommonScraper = ScrapePerformanceData{}

	clusterDeleteRequestPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "cluster_delete_request_per_second"),
		"集群 DELETE 请求次数",
		[]string{}, nil,
	)
	clusterGetRequestPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "cluster_get_request_per_second"),
		"集群 GET 请求次数",
		[]string{}, nil,
	)
	clusterPutRequestPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "cluster_put_request_per_second"),
		"集群 PUT 请求次数",
		[]string{}, nil,
	)
	clusterPostRequestPerSecond = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "cluster_post_request_per_second"),
		"集群 POST 请求次数",
		[]string{}, nil,
	)

	clusterReadBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "cluster_read_bandwidth"),
		"集群读带宽,KiB/s",
		[]string{}, nil,
	)
	clusterWriteBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "cluster_write_bandwidth"),
		"集群写带宽,KiB/s",
		[]string{}, nil,
	)
	clusterTotalBandwidth = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "cluster_total_bandwidth"),
		"集群总带宽,KiB/s",
		[]string{}, nil,
	)
)

// ScrapePerformanceData 是将要实现 Scraper 接口的一个 Metric 结构体
type ScrapePerformanceData struct{}

// Name 指定自己定义的 抓取器 的名字，与 Metric 的名字不是一个概念，但是一般保持一致
func (ScrapePerformanceData) Name() string {
	return "performance_data"
}

// Help 指定自己定义的 抓取器 的帮助信息，这里的 Help 的内容将会作为命令行标志的帮助信息。与 Metric 的 Help 不是一个概念。
func (ScrapePerformanceData) Help() string {
	return "HWObs Performance Data"
}

// Scrape 从客户端采集数据，并将其作为 Metric 通过 channel(通道) 发送。主要就是采集 HWObs 集群信息的具体行为。
func (ScrapePerformanceData) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	url := "/api/v2/pms/performance_data"

	// 配置请求体参数
	objectX := object{
		ObjectType: 57347,
		Indicators: []int{
			DELETERequestPerSecond,
			GETRequestPerSecond,
			PUTRequestPerSecond,
			POSTRequestPerSecond,
			ReadBandwidth,
			WriteBandwidth,
			TotalBandwidth,
		},
	}
	objects := append([]object{}, objectX)

	reqBody := reqBodyData{
		Objects:   objects,
		BeginTime: time.Now().UnixNano()/1e9 - 1000,
		EndTime:   time.Now().UnixNano()/1e9 - 990,
	}

	// 解析请求体
	reqBodyByte, _ := json.Marshal(reqBody)

	// 用于绑定响应体的结构体
	var (
		performanceDataRespBody []byte
		performanceData         performanceData
	)

	// 发起请求
	if performanceDataRespBody, err = client.Request("POST", url, bytes.NewBuffer(reqBodyByte)); err != nil {
		return err
	}
	if err = json.Unmarshal(performanceDataRespBody, &performanceData); err != nil || performanceData.Result.Code != 0 || len(performanceData.Data) < 1 {
		return err
	}

	// 创建一个新的 map，并以指标标识符作为 key，响应体中的 Data 字段内容作为 value 保存。
	// 主要是为了将 Data 数组中的元素进行分类，以便可以简单的输出监控指标。
	p := make(map[int]pData)
	for _, performance := range performanceData.Data {
		i, _ := strconv.Atoi(performance.Indicator)
		p[i] = performance
	}

	logrus.Debugf("性能数据响应信息：%v", p)

	// 集群 DELETE 请求次数
	deleteRequestPerSecond, _ := strconv.ParseFloat(p[DELETERequestPerSecond].IndicatorValues[0], 64)
	ch <- prometheus.MustNewConstMetric(clusterDeleteRequestPerSecond, prometheus.GaugeValue, deleteRequestPerSecond)
	// 集群 GET 请求次数
	getRequestPerSecond, _ := strconv.ParseFloat(p[GETRequestPerSecond].IndicatorValues[0], 64)
	ch <- prometheus.MustNewConstMetric(clusterGetRequestPerSecond, prometheus.GaugeValue, getRequestPerSecond)
	// 集群 PUT 请求次数
	putRequestPerSecond, _ := strconv.ParseFloat(p[PUTRequestPerSecond].IndicatorValues[0], 64)
	ch <- prometheus.MustNewConstMetric(clusterPutRequestPerSecond, prometheus.GaugeValue, putRequestPerSecond)
	// 集群 POST 请求次数
	postRequestPerSecond, _ := strconv.ParseFloat(p[POSTRequestPerSecond].IndicatorValues[0], 64)
	ch <- prometheus.MustNewConstMetric(clusterPostRequestPerSecond, prometheus.GaugeValue, postRequestPerSecond)
	// 集群读带宽
	readBandwidth, _ := strconv.ParseFloat(p[ReadBandwidth].IndicatorValues[0], 64)
	ch <- prometheus.MustNewConstMetric(clusterReadBandwidth, prometheus.GaugeValue, readBandwidth)
	// 集群写带宽
	writeBandwidth, _ := strconv.ParseFloat(p[WriteBandwidth].IndicatorValues[0], 64)
	ch <- prometheus.MustNewConstMetric(clusterWriteBandwidth, prometheus.GaugeValue, writeBandwidth)
	// 集群总带宽
	totalBandwidth, _ := strconv.ParseFloat(p[TotalBandwidth].IndicatorValues[0], 64)
	ch <- prometheus.MustNewConstMetric(clusterTotalBandwidth, prometheus.GaugeValue, totalBandwidth)

	return nil
}

// performanceData 性能数据
type performanceData struct {
	Data   []pData `json:"data"`
	Result pResult `json:"result"`
}
type pData struct {
	ID              string   `json:"id"`
	Indicator       string   `json:"indicator"`
	IndicatorValues []string `json:"indicator_values"`
	Name            string   `json:"name"`
	ObjectType      string   `json:"object_type"`
	Timestamp       []int    `json:"timestamp"`
}
type pResult struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

// 请求体参数的结构体
type reqBodyData struct {
	//
	Objects   []object `json:"objects"`
	BeginTime int64    `json:"begin_time"`
	EndTime   int64    `json:"end_time"`
}
type object struct {
	ObjectType int   `json:"object_type"`
	Indicators []int `json:"indicators"`
}
