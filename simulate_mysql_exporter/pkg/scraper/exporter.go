package scraper

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	name      = "common_exporter"
	namespace = "common"
	//Subsystem(s).
	exporter = "exporter"
)

// 这个指标用来统计，程序采集其他指标时所花费的时间
var (
	ScrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, exporter, "collector_duration_seconds"),
		"Collector time duration.",
		[]string{"collector"}, nil,
	)
)

// Metrics 本程序默认自带的一些 Metrics
type Metrics struct {
	TotalScrapes prometheus.Counter
	ScrapeErrors *prometheus.CounterVec
	Error        prometheus.Gauge
	XskyUP       prometheus.Gauge
}

// NewMetrics 实例化 Metrics，设定本程序默认自带的一些 Metrics 的信息
func NewMetrics() Metrics {
	subsystem := exporter
	return Metrics{
		TotalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "scrapes_total",
			Help:      "Total number of times Xsky was scraped for metrics.",
		}),
		ScrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occurred scraping a Xsky.",
		}, []string{"collector"}),
		Error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from Xsky resulted in an error (1 for error, 0 for success).",
		}),
		XskyUP: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Whether the Xsky is up.",
		}),
	}
}

// Exporter 实现了 prometheus.Collector，其中包含了很多 Metric。
// 只要 Exporter 实现了 prometheus.Collector，就可以调用 MustRegister() 将其注册到 prometheus 库中
type Exporter struct {
	client   CommonClient
	scrapers []CommonScraper
	metrics  Metrics
}

// NewExporter 实例化 Exporter
func NewExporter(client CommonClient, metrics Metrics, scrapers []CommonScraper) (*Exporter, error) {
	return &Exporter{
		client:   client,
		metrics:  metrics,
		scrapers: scrapers,
	}, nil
}

// Describe 实现 Collector 接口的方法
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.metrics.TotalScrapes.Desc()
	e.metrics.ScrapeErrors.Describe(ch)
	ch <- e.metrics.Error.Desc()
	ch <- e.metrics.XskyUP.Desc()
}

// Collect 实现 Collector 接口的方法
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	// 将 scrape() 方法引进来，用来在实现 Collect 接口后，调用 prometheus 功能可以操作 scrape() 中相关的 Metrics
	e.scrape(ch)

	ch <- e.metrics.TotalScrapes
	e.metrics.ScrapeErrors.Collect(ch)
	ch <- e.metrics.Error
	ch <- e.metrics.XskyUP
}

// scrape 调用每个已经注册的 Scraper(抓取器) 执行其代码中定义的抓取行为。
func (e *Exporter) scrape(ch chan<- prometheus.Metric) {
	// 每执行一次 scrape，TotalScraple 这个 Metrci 的值加一，用于统计从启动到现在采集了多少次
	e.metrics.TotalScrapes.Inc()

	// 第一个 scrapeTime,开始统计 scrape 指标的耗时
	scrapeTime := time.Now()

	// 检验目标服务器是否正常，每次执行 Collect 都会检查
	// 然后为 XskyUP 和 Error 这俩 Metrics 设置值。
	// if pong, err := e.client.Ping(); pong != true || err != nil {
	// 	log.WithFields(log.Fields{
	// 		"url": e.client.Opts.URL + "/health",
	// 	}).Error(err)
	// 	e.metrics.XskyUP.Set(0)
	// 	e.metrics.Error.Set(1)
	// }
	e.metrics.XskyUP.Set(1)
	e.metrics.Error.Set(0)

	// 对应第一个 scrapeTime，显示 scrapeDurationDesc 这个 Metric 的标签为 reach 的时间。也就是检验目标服务器状态总共花了多长时间
	ch <- prometheus.MustNewConstMetric(ScrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeTime).Seconds(), "reach")

	var wg sync.WaitGroup
	defer wg.Wait()

	// ！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！
	// 本代码中最核心的执行部分，通过一个 for 循环来执行所有经注册的 Scraper
	// ！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！
	// 由于所有自定义的 Scrapers 都实现了 Scraper 接口，所以这里的 e.scrapers 其实是那些 抓取器 结构体的集合
	for _, s := range e.scrapers {
		wg.Add(1)
		// go 协程，同时执行所有 Scraper
		go func(s CommonScraper) {
			defer wg.Done()
			// 第二个 scrapeTime,开始统计 scrape 指标的耗时
			label := s.Name()
			scrapeTime := time.Now()
			// 执行 Scrape 操作，也就是执行每个 Scraper 中的 Scrape() 方法，由于这些自定义的 Scraper 都实现了 Scraper 接口
			// 所以 Scrape 这个调用，就是调用的当前循环体中，从 e.scrapers 数组中取到的值，也就是 collector.ScrapeCluster{} 这些结构体
			if err := s.Scrape(e.client, ch); err != nil {
				logrus.WithField("scraper", s.Name()).Error(err)
				e.metrics.ScrapeErrors.WithLabelValues(label).Inc()
				e.metrics.Error.Set(1)
			}
			// 对应第二个 scrapeTime，scrapeDurationDesc 这个 Metric，用于显示抓取标签为 label(这是变量) 指标所消耗的时间
			// 其实就是统计每个 Scraper 执行所消耗的时间
			ch <- prometheus.MustNewConstMetric(ScrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeTime).Seconds(), label)
		}(s)
	}
}
