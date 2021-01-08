package scraper

import (
	"github.com/prometheus/client_golang/prometheus"
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
