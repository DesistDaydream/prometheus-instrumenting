package collector

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const (
	name      = "xsky_exporter"
	namespace = "xsky"
	//Subsystem(s).
	exporter = "exporter"
)

// Name is
func Name() string {
	return name
}

// Verify if Exporter implements prometheus.Collector
var _ prometheus.Collector = (*Exporter)(nil)

// Metric descriptors.
var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, exporter, "collector_duration_seconds"),
		"Collector time duration.",
		[]string{"collector"}, nil,
	)
)

// Exporter 实现了 prometheus.Collector，其中包含了很多 Metric。
// 只要 Exporter 实现了 prometheus.Collector，就可以调用 MustRegister() 将其注册到 prometheus 库中
type Exporter struct {
	//ctx      context.Context  //http timeout will work, don't need this
	client   *XskyClient
	scrapers []Scraper
	metrics  Metrics
}

// NewExporter 实例化 Exporter
func NewExporter(opts *XskyOpts, metrics Metrics, scrapers []Scraper) (*Exporter, error) {
	uri := opts.URL
	if !strings.Contains(uri, "://") {
		uri = "http://" + uri
	}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid Xsky URL: %s", err)
	}
	if u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("invalid Xsky URL: %s", uri)
	}

	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	tlsClientConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    rootCAs,
	}

	if opts.Insecure {
		tlsClientConfig.InsecureSkipVerify = true
	}

	transport := &http.Transport{
		TLSClientConfig: tlsClientConfig,
	}

	xc := &XskyClient{
		Opts: opts,
		Client: &http.Client{
			Timeout:   opts.Timeout,
			Transport: transport,
		},
	}

	return &Exporter{
		client:   xc,
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

func (e *Exporter) scrape(ch chan<- prometheus.Metric) {
	e.metrics.TotalScrapes.Inc()

	// 第一个 scrapeTime,开始统计 scrape 指标的耗时
	scrapeTime := time.Now()

	// 检验目标服务器是否正常，每次执行 Collect 都会检查
	if pong, err := e.client.Ping(); pong != true || err != nil {
		log.WithFields(log.Fields{
			"url":      e.client.Opts.URL + "/configurations",
			"username": e.client.Opts.Username,
		}).Error(err)
		e.metrics.XskyUP.Set(0)
		e.metrics.Error.Set(1)
	}
	e.metrics.XskyUP.Set(1)
	e.metrics.Error.Set(0)

	// 对应第一个 scrapeTime，scrapeDurationDesc 这个 Metric 用于显示抓取标签为 reach 指标所消耗的时间
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeTime).Seconds(), "reach")

	var wg sync.WaitGroup
	defer wg.Wait()
	for _, scraper := range e.scrapers {
		wg.Add(1)
		go func(scraper Scraper) {
			defer wg.Done()
			// 第二个 scrapeTime,开始统计 scrape 指标的耗时
			label := scraper.Name()
			scrapeTime := time.Now()
			if err := scraper.Scrape(e.client, ch); err != nil {
				log.WithField("scraper", scraper.Name()).Error(err)
				e.metrics.ScrapeErrors.WithLabelValues(label).Inc()
				e.metrics.Error.Set(1)
			}
			// 对应第二个 scrapeTime，scrapeDurationDesc 这个 Metric，用于显示抓取标签为 label(这是变量) 指标所消耗的时间
			ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeTime).Seconds(), label)
		}(scraper)
	}
}

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
