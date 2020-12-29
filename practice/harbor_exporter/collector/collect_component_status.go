package collector

import (
	simplejson "github.com/bitly/go-simplejson"
)

// HarborComponentStatue is
type HarborComponentStatue struct {
	ComponentNames  []string
	ComponentStatus []float64
}

// HarborHealthCollector 获取 Harbor 各组件健康状态
func HarborHealthCollector(h *HarborConnInfo) *HarborComponentStatue {
	var names []string
	var status []float64
	var statue float64

	// 使用指定的 endpoint 从 Harbor API 中获取信息
	body, _ := h.HarborConn("/health")
	jsonBody, _ := simplejson.NewJson(body)

	// 获取全部组件的总体健康状态
	names = append(names, "all")
	statueString, _ := jsonBody.Get("status").String()
	if statueString == "healthy" {
		statue = 1
	} else {
		statue = 0
	}
	status = append(status, statue)

	// 获取各组件单独的健康状态
	for i := 0; i < len(jsonBody.Get("components").MustArray()); i++ {
		// 逐一获取组件名称
		componentName, _ := jsonBody.Get("components").GetIndex(i).Get("name").String()
		names = append(names, componentName)

		// 逐一获取组件状态
		statueString, _ := jsonBody.Get("components").GetIndex(i).Get("status").String()
		if statueString == "healthy" {
			statue = 1
		} else {
			statue = 0
		}
		status = append(status, statue)

		// fmt.Println(hcs.ComponentNames, hcs.ComponentStatus)
	}

	return &HarborComponentStatue{
		ComponentNames:  names,
		ComponentStatus: status,
	}
}
