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
func (hcs *HarborComponentStatue) HarborHealthCollector(h *HarborConnInfo) {
	var componentStatue float64
	// 使用指定的 endpoint 从 Harbor API 中获取信息
	body, _ := h.HarborConn("/health")
	jsonBody, _ := simplejson.NewJson(body)

	// 获取其余组件的健康状态
	for i := 0; i < len(jsonBody.Get("components").MustArray()); i++ {
		//
		componentName, _ := jsonBody.Get("components").GetIndex(i).Get("name").String()
		hcs.ComponentNames = append(hcs.ComponentNames, componentName)

		//
		componentStatusValue, _ := jsonBody.Get("components").GetIndex(i).Get("status").String()
		if componentStatusValue == "healthy" {
			componentStatue = 1
		} else {
			componentStatue = 0
		}
		hcs.ComponentStatus = append(hcs.ComponentStatus, componentStatue)

		// fmt.Println(hcs.ComponentNames, hcs.ComponentStatus)
	}
}
