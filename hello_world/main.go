package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// promhttp.Handler() 函数旨再涵盖大部分基本用例。
	// 如果需要更多自定义的操作（包括使用非默认 Gatherer、不同的检测和非默认 HandlerOpts），使用 HandlerFor() 函数。请参阅那里了解详细信息。
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
