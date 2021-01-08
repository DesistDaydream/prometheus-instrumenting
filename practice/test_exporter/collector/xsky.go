package collector

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

// XskyClient 连接 Xsky 所需信息
type XskyClient struct {
	Client *http.Client
	Token  string
	Opts   *XskyOpts
}

// XskyOpts 登录 Xsky 所需属性
type XskyOpts struct {
	URL      string
	Username string
	password string
	// 这俩是关于 http.Client 的选项
	Timeout  time.Duration
	Insecure bool
}

// AddFlag use after set Opts
func (o *XskyOpts) AddFlag() {
	flag.StringVar(&o.URL, "xsky-server", "http://10.20.5.98:8056", "HTTP API address of a harbor server or agent. (prefix with https:// to connect over HTTPS)")
	flag.StringVar(&o.Username, "xsky-user", "admin", "xsky username")
	flag.StringVar(&o.password, "xsky-pass", "admin", "xsky password")
	flag.DurationVar(&o.Timeout, "time-out", time.Millisecond*1600, "Timeout on HTTP requests to the harbor API.")
	flag.BoolVar(&o.Insecure, "insecure", true, "Disable TLS host verification.")
}

// GetToken 获取 Xsky 认证所需 Token
func (x *XskyClient) GetToken() (err error) {
	// 设置 json 格式的 request body
	jsonReqBody := []byte("{\"auth\":{\"name\":\"" + x.Opts.Username + "\",\"password\":\"" + x.Opts.password + "\"}}")
	// 设置 URL
	url := fmt.Sprintf("%v/api/v1/auth/tokens:login", x.Opts.URL)
	// 设置 Request 信息
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonReqBody))
	req.Header.Add("Content-Type", "application/json")

	// 发送 Request 并获取 Response
	resp, err := x.Client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 处理 Response Body,并获取 Token
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	jsonRespBody, err := simplejson.NewJson(respBody)
	if err != nil {
		return
	}
	// fmt.Printf("本次响应的 Body 为：%v\n", string(respBody))
	x.Token, _ = jsonRespBody.Get("token").Get("uuid").String()
	fmt.Println("成功获取 Token！ ", x.Token)
	return
}

// Request 建立与 Xsky 的连接，并返回 Response Body
func (x *XskyClient) Request(endpoint string) (body []byte, err error) {
	// 获取 Xsky 认证所需 Token
	// TODO 还需要添加一个认证，当 Token 失效时，也需要重新获取 Token
	if x.Token == "" {
		x.GetToken()
	}
	fmt.Println("Xsky Token 为：", x.Token)

	// 根据认证信息及 endpoint 参数，创建与 Xsky 的连接，并返回 Body 给每个 Metric 采集器
	var resp *http.Response
	url := x.Opts.URL + endpoint
	log.Debugf("request url %s", url)

	// 创建一个新的 Request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(x.Opts.Username, x.Opts.password)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Cookie", "XMS_AUTH_TOKEN="+x.Token)

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

// Ping 在 Scraper 接口的实现方法 scrape() 中调用。
// 让 Exporter 每次获取数据时，都检验一下目标设备通信是否正常
func (x *XskyClient) Ping() (bool, error) {
	// fmt.Println("每次从 Xsky 获取数据时，都会进行测试")
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
