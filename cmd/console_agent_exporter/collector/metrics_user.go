package collector

import (
	"bytes"
	"encoding/json"

	"github.com/DesistDaydream/prometheus-instrumenting/pkg/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ scraper.CommonScraper = ScrapeUser{}

	users = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "gdas_user_count"),
		"集群中所有用户的总数",
		[]string{}, nil,
	)
)

// ScrapeUser 抓取用户信息
type ScrapeUser struct{}

// Name is
func (ScrapeUser) Name() string {
	return "gdas_user_count"
}

// Help is
func (ScrapeUser) Help() string {
	return "Gdas User Info"
}

// Scrape is
func (ScrapeUser) Scrape(client scraper.CommonClient, ch chan<- prometheus.Metric) (err error) {
	var userdata userdata

	url := "/api/gdas/user/list"
	method := "POST"
	reqBody := `{"pages":"1-1", "userName":""}`
	buf := bytes.NewBuffer([]byte(reqBody))
	respBody, err := client.Request(method, url, buf)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &userdata)
	if err != nil {
		return err
	}

	userCount := float64(userdata.ResCount)
	ch <- prometheus.MustNewConstMetric(users, prometheus.GaugeValue, userCount)
	return nil
}

type userdata struct {
	Result   string     `json:"result"`
	ResCount int        `json:"res_count"`
	UserList []userList `json:"userList"`
}

type userList struct {
	UserName         string `json:"userName"`
	UserAuth         int    `json:"userAuth"`
	Ak               string `json:"ak"`
	Sk               string `json:"sk"`
	Active           bool   `json:"active"`
	UserLoginChk     bool   `json:"userLoginChk"`
	UserLoginChkTime int    `json:"userLoginChkTime"`
}
