package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/DesistDaydream/prometheus-instrumenting/exporter/huawei_obs_exporter/collector"
	"github.com/DesistDaydream/prometheus-instrumenting/exporter/pkg/scraper"
	"github.com/coreos/go-systemd/daemon"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var scrapers = map[scraper.CommonScraper]bool{
	collector.ScrapeCluster{}:         true,
	collector.ScrapeDisk{}:            true,
	collector.ScrapeStoragePool{}:     true,
	collector.ScrapePerformanceData{}: true,
}

func LogInit(level, file string) error {
	// logrus.SetFormatter(&logrus.TextFormatter{
	// 	FullTimestamp:   true,
	// 	TimestampFormat: "2006-01-02 15:04:05",
	// })
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat:   "2006-01-02 15:04:05",
		DisableTimestamp:  false,
		DisableHTMLEscape: false,
		DataKey:           "",
		// FieldMap:          map[logrus.fieldKey]string{},
		// CallerPrettyfier: func(*runtime.Frame) (string, string) {},
		PrettyPrint: false,
	})
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(logLevel)

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
	scraper.Namespace = collector.Namespace

	listenAddress := pflag.String("web.listen-address", ":9122", "Address to listen on for web interface and telemetry.")
	metricsPath := pflag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	logLevel := pflag.String("log-level", "info", "The logging level:[debug, info, warn, error, fatal]")
	logFile := pflag.String("log-output", "", "the file which log to, default stdout")

	opts := &collector.HWObsOpts{}
	opts.AddFlag()

	scraperFlags := map[scraper.CommonScraper]*bool{}
	for scraper, enabledByDefault := range scrapers {
		defaultOn := false
		if enabledByDefault {
			defaultOn = true
		}
		f := pflag.Bool("collect."+scraper.Name(), defaultOn, scraper.Help())
		scraperFlags[scraper] = f
	}
	// 解析命令行标志,即：将命令行标志的值传递到代码的变量中。若不解析，则所有通过命令行标志设置的变量是没有值的。
	pflag.Parse()

	// 初始化日志
	if err := LogInit(*logLevel, *logFile); err != nil {
		logrus.Fatal(errors.Wrap(err, "set log level error"))
	}

	// ######## 下面的都是 Exporter 运行的最主要逻辑了 ########
	enabledScrapers := []scraper.CommonScraper{}
	for scraper, enabled := range scraperFlags {
		if *enabled {
			logrus.Info("Scraper enabled ", scraper.Name())
			enabledScrapers = append(enabledScrapers, scraper)
		}
	}
	hwObsClient := collector.NewHWObsClient(opts)
	exporter := scraper.NewExporter(hwObsClient, enabledScrapers)
	reg := prometheus.NewRegistry()
	reg.MustRegister(exporter)

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

	logrus.Info("Listening on address ", *listenAddress)
	daemon.SdNotify(false, daemon.SdNotifyReady)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		logrus.Fatal(err)
	}
}
