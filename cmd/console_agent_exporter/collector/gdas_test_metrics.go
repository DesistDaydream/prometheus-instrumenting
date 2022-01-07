package collector

import (
	"fmt"

	"github.com/DesistDaydream/prometheus-instrumenting/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ scraper.CommonScraper = ScrapeTestMetrics{}

	testMetrics = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_test_metrics"),
		"Test",
		[]string{}, nil,
	)
)

// ScrapeTestMetrics is
type ScrapeTestMetrics struct{}

// Name is
func (ScrapeTestMetrics) Name() string {
	return "gdas_test_metrics"
}

// Help is
func (ScrapeTestMetrics) Help() string {
	return "Gdas Temporary test"
}

// Scrape is
func (ScrapeTestMetrics) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	var (
		respBody []byte
		// testdata testdata
	)

	url := "/api/gdas/magazines"
	method := "GET"
	if respBody, err = client.Request(method, url, nil); err != nil {
		return err
	}

	fmt.Printf("功能测试，响应体为：%v\n", string(respBody))

	return nil
}
