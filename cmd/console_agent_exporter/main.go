package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/DesistDaydream/prometheus-instrumenting/cmd/console_agent_exporter/collector"
	"github.com/DesistDaydream/prometheus-instrumenting/pkg/scraper"
	"github.com/coreos/go-systemd/daemon"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// scrapers 列出了应该注册的所有 Scraper(抓取器)，以及默认情况下是否应该启用它们
// 用一个 map 来定义这些抓取器是否开启，key 为 collector.Scraper 接口类型，value 为 bool 类型。
// 凡是实现了 collector.Scraper 接口的结构体，都可以做作为该接口类型的值
var scrapers = map[scraper.CommonScraper]bool{
	collector.ScrapeDas{}:              true,
	collector.ScrapePool{}:             true,
	collector.ScrapeUser{}:             true,
	collector.ScrapeTotalspace{}:       true,
	collector.ScrapeNodes{}:            true,
	collector.ScrapeMagazinesMetrics{}: true,
	// collector.ScrapeTestMetrics{}:      true,
	// ScrapeGc{}:          false,
	// ScrapeRegistries{}:  false,
}

// DumpStacks is
func DumpStacks() {
	buf := make([]byte, 16384)
	buf = buf[:runtime.Stack(buf, true)]
	logrus.Printf("=== BEGIN goroutine stack dump ===\n%s\n=== END goroutine stack dump ===", buf)
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
	// 设置通用包中的指标的前缀
	scraper.Namespace = collector.Namespace

	// 设置命令行标志，开始
	//
	listenAddress := pflag.String("web.listen-address", ":19097", "Address to listen on for web interface and telemetry.")
	metricsPath := pflag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	logLevel := pflag.String("log-level", "info", "The logging level:[debug, info, warn, error, fatal]")
	logFile := pflag.String("log-output", "", "the file which log to, default stdout")
	// pflag.StringVar(&collector.HarborVersion, "override-version", "", "override the harbor version")

	// 设置关于抓取 Metric 目标客户端的一些信息的标志
	opts := &collector.ConsoleAgentOpts{}
	opts.AddFlag()

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
	//
	// 设置命令行标志，结束

	// 初始化日志
	if err := LogInit(*logLevel, *logFile); err != nil {
		logrus.Fatal(errors.Wrap(err, "set log level error"))
	}

	// 下面的都是 Exporter 运行的最主要逻辑了
	//
	// 获取所有通过命令行标志，设置开启的 scrapers(抓取器)。
	// 不包含默认开启的，默认开启的在代码中已经指定了。
	enabledScrapers := []scraper.CommonScraper{}
	for scraper, enabled := range scraperFlags {
		if *enabled {
			logrus.Info("Scraper enabled ", scraper.Name())
			enabledScrapers = append(enabledScrapers, scraper)
		}
	}

	// 实例化 Exporter，其中包括所有自定义的 Metrics
	consoleAgentClient := collector.NewConsoleAgentClient(opts)
	exporter := scraper.NewExporter(consoleAgentClient, enabledScrapers)
	// 实例化一个注册器,并使用这个注册器注册 exporter
	reg := prometheus.NewRegistry()
	reg.MustRegister(exporter)

	// 设置路由信息
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

	// 启动前检查并启动 Exporter
	logrus.Info("Listening on address ", *listenAddress)
	daemon.SdNotify(false, daemon.SdNotifyReady)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		logrus.Fatal(err)
	}
}
