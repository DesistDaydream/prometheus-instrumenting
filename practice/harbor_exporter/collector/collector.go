package collector

import (
	"fmt"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/prometheus/client_golang/prometheus"
)

// HarborHealthCollector 获取 Harbor 各组件健康状态
func HarborHealthCollector(h *HarborConnInfo) {
	var healthstatus float64
	// 使用指定的 endpoint 从 Harbor API 中获取信息
	body, _ := h.HarborConn("/health")
	jsonBody, _ := simplejson.NewJson(body)
	health, _ := jsonBody.Get("status").String()
	if health == "healthy" {
		healthstatus = 0
	} else {
		healthstatus = 1
	}
	// 为指标设置值
	// HarborHealthStatus.Set(healthstatus)
	HarborHealthStatus.With(prometheus.Labels{"component": "all"}).Set(healthstatus)

	// 获取其余组件的健康状态
	for i := 0; i < len(jsonBody.Get("components").MustArray()); i++ {
		componentName, _ := jsonBody.Get("components").GetIndex(i).Get("name").String()
		componentStatus, _ := jsonBody.Get("components").GetIndex(i).Get("status").String()
		if componentStatus == "healthy" {
			healthstatus = 0
		} else {
			healthstatus = 1
		}
		fmt.Println(componentName, componentStatus)
		HarborHealthStatus.With(prometheus.Labels{"component": componentName}).Set(healthstatus)
	}
}
