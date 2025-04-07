package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

var ctx = context.Background()

// 示例 1: 即时向量查询
func InstantVectorQuery(promAPI v1.API) {
	result, warnings, err := promAPI.Query(ctx, "up", time.Now())
	if err != nil {
		log.Fatalf("执行查询错误: %v", err)
	}
	if len(warnings) > 0 {
		fmt.Printf("查询警告: %v\n", warnings)
	}

	fmt.Println("######## 即时查询结果 ########")
	fmt.Printf("类型: %T\n", result)
	fmt.Println(result)
}

// 示例 2: 范围向量查询
func RangeVectorQuery(promAPI v1.API, startTime time.Time, endTime time.Time, step time.Duration) {
	rangeResult, warnings, err := promAPI.QueryRange(ctx, `rate(node_cpu_seconds_total{mode="system",cpu="0"}[5m])`, v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  step,
	})
	if err != nil {
		log.Fatalf("执行范围查询错误: %v", err)
	}
	if len(warnings) > 0 {
		fmt.Printf("范围查询警告: %v\n", warnings)
	}

	fmt.Printf("类型: %T\n", rangeResult)
	fmt.Println("######## 范围查询结果 ########")
	if matrix, ok := rangeResult.(model.Matrix); ok {
		for _, sampleStream := range matrix {
			// 指标的 Labels 集
			fmt.Println(sampleStream.Metric)
			// 指标的样本值和样本时间戳
			for _, value := range sampleStream.Values {
				fmt.Println(value)
			}
		}
	}
}

// 示例 3: 查询指标标签
func LabelNamesQuery(promAPI v1.API, startTime time.Time, endTime time.Time) {
	labelNames, warnings, err := promAPI.LabelNames(ctx, []string{`up`}, startTime, endTime)
	if err != nil {
		log.Fatalf("查询标签名称错误: %v", err)
	}
	if len(warnings) > 0 {
		fmt.Printf("标签查询警告: %v\n", warnings)
	}

	fmt.Println("######## 标签名称列表查询结果 ########")
	fmt.Println(labelNames)
}

func metaDataQuery(promAPI v1.API) {
	metadata, err := promAPI.Metadata(ctx, "", "")
	if err != nil {
		log.Fatalf("查询指标名称错误: %v", err)
	}

	fmt.Println("######## 列出所有指标名称 ########")
	for metric := range metadata {
		fmt.Printf("- %s\n", metric)
	}
}

func main() {
	// 创建 Prometheus API 客户端配置
	config := api.Config{
		Address: "http://192.168.254.253:9090", // Prometheus 服务器地址，根据实际情况修改
	}

	// 实例化 Prom 客户端
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("创建 Prometheus 客户端错误: %v", err)
	}

	// 实例化 Prom API
	promAPI := v1.NewAPI(client)

	// 设定查询时间
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour) // 查询过去1小时的数据
	step := time.Minute * 20                 // 步长

	InstantVectorQuery(promAPI)

	RangeVectorQuery(promAPI, startTime, endTime, step)

	LabelNamesQuery(promAPI, startTime, endTime)

	metaDataQuery(promAPI)

	a, _ := promAPI.Alerts(ctx)
	fmt.Println(a)
}
