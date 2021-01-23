package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/DesistDaydream/exporter/simulate_mysql_exporter/pkg/scraper"
	"github.com/DesistDaydream/exporter/simulate_mysql_exporter/xsky_exporter/collector"
	"github.com/coreos/go-systemd/daemon"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

/*
该项目起源于 mysql exporter，是对 mysql exporter 的改装
该项目对 prometheus.Collector 接口进行了二次封装，通过一个名为 Scraper 的接口中的 Scrape() 方法，来定义具体的采集行为
scraper.go 定义了 Scraper 接口
exporter.go 文件定义了一个实现了 Scraper 接口的结构体。这个结构体包含两种属性。1.所有 Metrics、 2.某个待采集目标的连接信息的
xsky.go 文件则是定义了连接 xsky 的方式，在 mysql expoter 项目中，这个文件叫 mysql.go,就是定义 mysql 的连接方式
metrics_XXXX.go 文件就是定义各种待采集的 Metrics，以及采集这些 Metrics 的具体行为。一个 Metric 放在一个文件中。
metrics_XXXX.go 文件中，包含了实现了 Scraper 接口的结构体。
*/

// scrapers 列出了应该注册的所有 Scraper(抓取器)，以及默认情况下是否应该启用它们
// 用一个 map 来定义这些抓取器是否开启，key 为 collector.Scraper 接口类型，value 为 bool 类型。
// 凡是实现了 collector.Scraper 接口的结构体，都可以做作为该接口类型的值
var scrapers = map[scraper.CommonScraper]bool{
	collector.ScrapeCluster{}: true,
	collector.ScrapeDisk{}:    true,
	collector.ScrapeDisk2{}:   true,
	// ScrapeGc{}:          false,
	// ScrapeRegistries{}:  false,
}

// LogInit 日志功能初始化，若指定了 log-output 命令行标志，则将日志写入到文件中
func LogInit(level, file string) error {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	le, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(le)

	if file != "" {
		f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		logrus.SetOutput(f)
	}

	return nil
}

func main() {
	// ####################################
	// ######## 设置命令行标志，开始 ########
	// ####################################
	listenAddress := pflag.String("web.listen-address", ":8080", "Address to listen on for web interface and telemetry.")
	metricsPath := pflag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	logLevel := pflag.String("log-level", "info", "The logging level:[debug, info, warn, error, fatal]")
	logFile := pflag.String("log-output", "", "the file which log to, default stdout")
	// pflag.StringVar(&collector.HarborVersion, "override-version", "", "override the harbor version")

	// 设置关于抓取 Metric 目标客户端的一些信息的标志
	opts := &collector.XskyOpts{}
	opts.AddFlag()
	pflag.Parse() // 由于在 NewXskyClient() 时，需要用到命令行标志的值，所以在此先解析一下，后面还需要再解析一次以识别其他的命令行标志的值
	xskyClient := collector.NewXsykClient(opts)

	// scraperFlags 也是一个 map，并且 key 为 collector.Scraper 接口类型，这一小段代码主要有下面几个作用
	// 1.生成抓取器的命令行标志，用于通过命令行控制开启哪些抓取器，说白了就是控制采集哪些指标
	// 2.下面的 for 循环会通过命令行 flag 获取到的值，放到 scraperFlags 这个 map 中
	// 3.然后在后面注册 Exporter 之前，先通过这个 map 中的键值对判断是否要把 value 为 true 的 抓取器 注册进去
	scraperFlags := map[scraper.CommonScraper]*bool{}
	for scraper, enabledByDefault := range scrapers {
		defaultOn := false
		if enabledByDefault {
			defaultOn = true
		}
		// 设置命令行 flag
		f := pflag.Bool("collect."+scraper.Name(), defaultOn, scraper.Help())
		// 将命令行 flag 中获取到的值，赋到 map 中，作为 map 的 value
		scraperFlags[scraper] = f
	}
	// 解析命令行标志
	pflag.Parse()
	// ####################################
	// ######## 设置命令行标志，结束 ########
	// ####################################

	// 初始化日志
	if err := LogInit(*logLevel, *logFile); err != nil {
		logrus.Fatal(errors.Wrap(err, "set log level error"))
	}

	// ######## 下面的都是 Exporter 运行的最主要逻辑了 ########
	// 获取所有通过命令行标志，设置开启的 scrapers(抓取器)。
	// 不包含默认开启的，默认开启的在代码中已经指定了。
	enabledScrapers := []scraper.CommonScraper{}
	for scraper, enabled := range scraperFlags {
		if *enabled {
			logrus.Info("Scraper enabled ", scraper.Name())
			enabledScrapers = append(enabledScrapers, scraper)
		}
	}
	// 实例化 Exporter，其中包括所有自定义的 Metrics。这里与 prometheus.Register() 的逻辑基本一致。
	// NewExporter 的两个接口分别用来传递 连接Server的信息 以及 需要采集的Metrics
	// 并且 NewExporter 返回的 Exporter 结构体，已经实现了 prometheus.Collector
	exporter := scraper.NewExporter(xskyClient, enabledScrapers)
	// 实例化一个注册器,并使用这个注册器注册 exporter
	reg := prometheus.NewRegistry()
	reg.MustRegister(exporter)
	// ######## Exporter 主要运行逻辑结束 ########

	// ######## 设置路由信息 ########
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>` + collector.Name() + `</title></head>
             <body>
             <h1>` + collector.Name() + `</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	http.Handle(*metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{ErrorLog: logrus.StandardLogger()}))
	http.HandleFunc("/-/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ok")
	})
	// ######## 设置路由信息结束 ########

	// 启动前检查并启动 Exporter
	logrus.Info("Listening on address ", *listenAddress)
	daemon.SdNotify(false, daemon.SdNotifyReady)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		logrus.Fatal(err)
	}
}
