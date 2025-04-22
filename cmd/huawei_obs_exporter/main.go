package main

import (
	"fmt"
	"net/http"

	logging "github.com/DesistDaydream/logging/pkg/logrus_init"

	"github.com/DesistDaydream/prometheus-instrumenting/cmd/huawei_obs_exporter/collector"
	"github.com/DesistDaydream/prometheus-instrumenting/pkg/scraper"
	"github.com/coreos/go-systemd/daemon"
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

var (
	logFlags logging.LogrusFlags
)

func main() {
	scraper.Namespace = collector.Namespace

	listenAddress := pflag.String("web.listen-address", ":18088", "Address to listen on for web interface and telemetry.")
	metricsPath := pflag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")

	logging.AddFlags(&logFlags)

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
	pflag.Parse()

	// 初始化日志
	if err := logging.LogrusInit(&logFlags); err != nil {
		logrus.Fatal("初始化日志失败", err)
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
