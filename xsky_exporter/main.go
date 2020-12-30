package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/DesistDaydream/exporter/xsky_exporter/collector"
	"github.com/coreos/go-systemd/daemon"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

var (
	version      string
	gitCommit    string
	gitTreeState = ""                     // state of git tree, either "clean" or "dirty"
	buildDate    = "1970-01-01T00:00:00Z" // build date, output of $(date +'%Y-%m-%dT%H:%M:%S')
)

// 在 / 页面输出的一些信息
func versionPrint() string {
	return fmt.Sprintf(`Name: %s
Version: %s
CommitID: %s
GitTreeState: %s
BuildDate: %s
GoVersion: %s
Compiler: %s
Platform: %s/%s
`, collector.Name(), version, gitCommit, gitTreeState, buildDate, runtime.Version(), runtime.Compiler, runtime.GOOS, runtime.GOARCH)
}

// setupSigusr1Trap is
func setupSigusr1Trap() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	go func() {
		for range c {
			DumpStacks()
		}
	}()
}

// DumpStacks is
func DumpStacks() {
	buf := make([]byte, 16384)
	buf = buf[:runtime.Stack(buf, true)]
	log.Printf("=== BEGIN goroutine stack dump ===\n%s\n=== END goroutine stack dump ===", buf)
}

// LogInit is
func LogInit(level, file string) error {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	le, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(le)

	if file != "" {
		f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		log.SetOutput(f)
	}

	return nil
}

func main() {
	// 设置命令行标志，开始
	//
	listenAddress := flag.String("web.listen-address", ":8080", "Address to listen on for web interface and telemetry.")
	metricsPath := flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	logLevel := flag.String("log-level", "info", "The logging level:[debug, info, warn, error, fatal]")
	logFile := flag.String("log-output", "", "the file which log to, default stdout")
	versionP := flag.Bool("version", false, "print version info")
	// flag.StringVar(&collector.HarborVersion, "override-version", "", "override the harbor version")

	// 设置关于抓取 Metric 目标客户端的一些信息的标志
	opts := &collector.XskyOpts{}
	opts.AddFlag()

	// 生成抓取器的命令行标志，用于通过命令行控制开启哪些抓取器
	// 说白了就是控制采集哪些指标
	scraperFlags := map[collector.Scraper]*bool{}
	for scraper, enabledByDefault := range collector.Scrapers {
		defaultOn := false
		if enabledByDefault {
			defaultOn = true
		}
		f := flag.Bool("collect."+scraper.Name(), defaultOn, scraper.Help())
		scraperFlags[scraper] = f
	}
	// 解析命令行标志
	flag.Parse()
	//
	// 设置命令行标志，结束

	// 通过 --version 命令行标志，可以获取 versionPrint() 函数中定义的信息
	if *versionP {
		fmt.Print(versionPrint())
		return
	}

	// 初始化日志
	if err := LogInit(*logLevel, *logFile); err != nil {
		log.Fatal(errors.Wrap(err, "set log level error"))
	}

	// 下面的都是 Exporter 运行的最主要逻辑了
	//
	// 获取所有通过命令行标志，设置开启的 scrapers(抓取器)。
	// 不包含默认开启的，默认开启的在代码中已经指定了。
	enabledScrapers := []collector.Scraper{}
	for scraper, enabled := range scraperFlags {
		if *enabled {
			log.Info("Scraper enabled ", scraper.Name())
			enabledScrapers = append(enabledScrapers, scraper)
		}
	}
	// 实例化 Exporter，其中包括所有自定义的 Metrics
	exporter, err := collector.NewExporter(opts, collector.NewMetrics(), enabledScrapers)
	if err != nil {
		log.Fatal(err)
	}
	// 实例化一个注册器,并使用这个注册器注册 exporter
	reg := prometheus.NewRegistry()
	reg.MustRegister(exporter)

	// 设置路由信息
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>` + collector.Name() + `</title></head>
             <body>
             <h1><a style="text-decoration:none" href='https://github.com/zhangguanzhang/harbor_exporter'>` + collector.Name() + `</a></h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             <h2>Build</h2>
             <pre>` + versionPrint() + `</pre>
             </body>
             </html>`))
	})

	http.Handle(*metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{ErrorLog: log.StandardLogger()}))

	http.HandleFunc("/-/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ok")
	})

	// 启动前检查并启动 Exporter
	setupSigusr1Trap()
	log.Info("Listening on address ", *listenAddress)
	daemon.SdNotify(false, daemon.SdNotifyReady)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatal(err)
	}

}
