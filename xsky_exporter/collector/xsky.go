package collector

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

var (
	// Scrapers is
	Scrapers = map[Scraper]bool{
		// ScrapeSystemInfo{}:  true,
		// ScrapeStatistics{}:  true,
		// ScrapeQuotas{}:      true,
		// ScrapeHealth{}:      true,
		// ScrapeProjects{}:    true,
		// ScrapeUsers{}:       true,
		// ScrapeLogs{}:        true,
		// ScrapeReplication{}: false,
		// ScrapeGc{}:          false,
		// ScrapeRegistries{}:  false,
	}

	// TODO
	//  tags always return full tag, see https://github.com/goharbor/harbor/issues/12279

	errResult = errors.New("cannot find data, maybe json is nil")
)

// XskyOpts 登录 Xsky 所需属性
type XskyOpts struct {
	URL      string
	Username string
	password string
	UA       string
	Timeout  time.Duration
	Insecure bool
}

// XskyClient 连接 Xsky 所需信息
type XskyClient struct {
	Client *http.Client
	Opts   *XskyOpts
}

// subInsJson could use for member and repos
type subInsJSON struct {
	ID        int `json:"id"`
	ProjectID int `json:"project_id"`
}

type idJSON struct {
	ID int `json:"id"`
}

// Error is
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// AddFlag use after set Opts
func (o *XskyOpts) AddFlag() {
	flag.StringVar(&o.URL, "xsky-server", "http://10.20.5.98:8056", "HTTP API address of a harbor server or agent. (prefix with https:// to connect over HTTPS)")
	flag.StringVar(&o.Username, "xsky-user", "admin", "xsky username")
	flag.StringVar(&o.password, "xsky-pass", "admin", "xsky password")
	flag.StringVar(&o.UA, "harbor-ua", "harbor_exporter", "user agent of the harbor http client")
	flag.DurationVar(&o.Timeout, "time-out", time.Millisecond*1600, "Timeout on HTTP requests to the harbor API.")
	flag.BoolVar(&o.Insecure, "insecure", false, "Disable TLS host verification.")
}

// Request 建立与 Xsky 的连接，并返回 Response Body
func (x *XskyClient) Request(endpoint string) (body []byte, err error) {
	var resp *http.Response
	url := x.Opts.URL + endpoint
	log.Debugf("request url %s", url)

	// 创建一个新的 Request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(x.Opts.Username, x.Opts.password)
	req.Header.Set("User-Agent", x.Opts.UA)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	// 根据新建立的 Request，发起请求，并获取 Response
	if resp, err = x.Client.Do(req); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error handling request for %s http-statuscode: %s", endpoint, resp.Status)
	}

	// 处理 Response Body
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}

	return body, nil
}

// Ping is
func (x *XskyClient) Ping() (bool, error) {
	req, err := http.NewRequest("GET", x.Opts.URL+"/configurations", nil)
	if err != nil {
		return false, err
	}
	req.SetBasicAuth(x.Opts.Username, x.Opts.password)

	resp, err := x.Client.Do(req)
	if err != nil {
		return false, err
	}

	resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
		return true, nil
	case resp.StatusCode == http.StatusUnauthorized:
		return false, errors.New("username or password incorrect")
	default:
		return false, fmt.Errorf("error handling request, http-statuscode: %s", resp.Status)
	}
}
