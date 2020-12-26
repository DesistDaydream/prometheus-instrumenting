package collector

import (
	simplejson "github.com/bitly/go-simplejson"
)

// HarborHealthGatherer 获取 Harbor 各组件健康状态
func HarborHealthGatherer(h *HarborConnInfo) {
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
	HarborHealthStatus.Set(healthstatus)
}
