package main

import (
	"context"
	"crypto/tls"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

var ctx = context.Background()

type PrometheusClient struct {
	URL string
	API v1.API
}

func NewPrometheusClient(url string) (*PrometheusClient, error) {
	client, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating Prometheus client: %v", err)
	}

	return &PrometheusClient{
		URL: url,
		API: v1.NewAPI(client),
	}, nil
}

func main() {
	// Define command line flags
	metricName := flag.String("metric", "hdf_hdf_version", "Metric name to fetch")
	outputFile := flag.String("output", "test/files/hdf_metrics.csv", "Output CSV file path")
	domain := flag.String("domain", "desistdaydream.it", "二级域名")
	operator := flag.String("operator", "mobile", "运营商")
	flag.Parse()

	allMetrics := make(map[string]map[string]string)
	allLabelNames := make(map[string]struct{})
	provinceList := map[string]string{"cq": "重庆", "fj": "福建", "gz": "广东", "hain": "海南", "nm": "内蒙", "sh": "上海", "sc": "四川", "xz": "西藏"}

	// Collect metrics from all Prometheus instances
	for key, value := range provinceList {
		url := fmt.Sprintf("https://prom.%s.%s.%s:10443", *operator, key, *domain)
		fmt.Printf("Fetching metrics from %s...\n", value)
		metrics, err := fetchMetrics(url, *metricName)
		if err != nil {
			fmt.Printf("Error fetching from %s: %v\n", url, err)
			continue
		}

		// Collect all metrics and build a set of all label names
		for fingerprint, labels := range metrics {
			allMetrics[fingerprint] = labels
			for labelName := range labels {
				allLabelNames[labelName] = struct{}{}
			}
		}

		fmt.Printf("Fetched %d time series from %s\n", len(metrics), url)
	}

	// Convert label names set to sorted slice for consistent CSV output
	labelNames := make([]string, 0, len(allLabelNames))
	for name := range allLabelNames {
		// Skip internal labels like __name__
		if !strings.HasPrefix(name, "__") {
			labelNames = append(labelNames, name)
		}
	}
	sort.Strings(labelNames)

	// Write to CSV
	if err := writeCSV(*outputFile, labelNames, allMetrics); err != nil {
		fmt.Printf("Error writing CSV: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully wrote %d time series to %s\n", len(allMetrics), *outputFile)
}

func fetchMetrics(url, metricName string) (map[string]map[string]string, error) {
	roundTripper := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// 创建 Prometheus API 客户端配置
	// 客户端使用 RoundTripper 来驱动 HTTP 请求。如果未提供，则将使用 DefaultRoundTripper
	config := api.Config{
		Address: url,
		RoundTripper: config.NewBasicAuthRoundTripper(
			config.NewInlineSecret("admin"),
			config.NewInlineSecret("haohan@Observ0"),
			roundTripper,
		),
	}

	// 实例化 Prom 客户端
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("创建 Prometheus 客户端错误: %v", err)
	}

	// 实例化 Prom API
	promAPI := v1.NewAPI(client)

	// 执行查询请求
	result, warnings, err := promAPI.Query(ctx, metricName, time.Now())
	if err != nil {
		return nil, fmt.Errorf("error querying Prometheus: %v", err)
	}

	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}

	vector, ok := result.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("expected vector result, got %T", result)
	}

	metrics := make(map[string]map[string]string)

	for _, sample := range vector {
		labels := make(map[string]string)
		for name, value := range sample.Metric {
			labels[string(name)] = string(value)
		}

		// Use fingerprint as a unique identifier
		metrics[sample.Metric.Fingerprint().String()] = labels
	}

	return metrics, nil
}

func writeCSV(filename string, labelNames []string, metrics map[string]map[string]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := append([]string{"fingerprint"}, labelNames...)
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %v", err)
	}

	// Write data rows
	for fingerprint, labels := range metrics {
		row := make([]string, len(header))
		row[0] = fingerprint

		for i, name := range labelNames {
			row[i+1] = labels[name]
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing row: %v", err)
		}
	}

	return nil
}
