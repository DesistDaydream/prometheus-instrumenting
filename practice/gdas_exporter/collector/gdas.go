package collector

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bitly/go-simplejson"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

// GdasOpts 登录 Gdas 所需属性
type GdasOpts struct {
	URL      string
	Username string
	password string
	Timeout  time.Duration
	Insecure bool
	Token    string
}

// GdasClient 连接 Gdas 所需信息
type GdasClient struct {
	Client *http.Client
	Opts   *GdasOpts
}

// AddFlag use after set Opts
func (o *GdasOpts) AddFlag() {
	flag.StringVar(&o.URL, "gdas-server", "https://172.38.30.192:8003", "HTTP API address of a harbor server or agent. (prefix with https:// to connect over HTTPS)")
	flag.StringVar(&o.Username, "gdas-user", "system", "gdas username")
	flag.StringVar(&o.password, "gdas-pass", "d153850931040e5c81e1c7508ded25f5f0ae76cb57dc1997bc343b878946ba23", "gdas password")
	flag.DurationVar(&o.Timeout, "time-out", time.Millisecond*1600, "Timeout on HTTP requests to the harbor API.")
	flag.BoolVar(&o.Insecure, "insecure", false, "Disable TLS host verification.")
}

// GetGdasToken 获取 Gdas 认证所需 Token
func (g *GdasClient) GetGdasToken() (err error) {
	// 设置 json 格式的 request body
	jsonReqBody := []byte("{\"userName\":\"" + g.Opts.Username + "\",\"passWord\":\"" + g.Opts.password + "\"}")
	// 设置 URL
	url := fmt.Sprintf("%v/v1/login", g.Opts.URL)
	// 设置 Request 信息
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonReqBody))
	req.Header.Add("referer", fmt.Sprintf("%v/v1/login", g.Opts.URL))
	req.Header.Add("Content-Type", "application/json")

	// 忽略证书验证
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 发送 Request 并获取 Response
	resp, err := (&http.Client{Transport: tr}).Do(req)
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
	g.Opts.Token, _ = jsonRespBody.Get("token").String()
	fmt.Println("成功获取 Token！ ", g.Opts.Token)
	return
}

// Request 建立与 Gdas 的连接，并返回 Response Body
func (g *GdasClient) Request(endpoint string) (body []byte, err error) {
	// 获取 Gdas 认证所需 Token
	// TODO 还需要添加一个认证，当 Token 失效时，也需要重新获取 Token
	if g.Opts.Token == "" {
		g.GetGdasToken()
	}
	fmt.Println("Gdas Token 为：", g.Opts.Token)

	// 根据认证信息及 endpoint 参数，创建与 Gdas 的连接，并返回 Body 给每个 Metric 采集器
	var resp *http.Response
	url := g.Opts.URL + endpoint
	log.Debugf("request url %s", url)

	// 创建一个新的 Request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(g.Opts.Username, g.Opts.password)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("token", g.Opts.Token)

	// 忽略证书验证
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 根据新建立的 Request，发起请求，并获取 Response
	if resp, err = (&http.Client{Transport: tr}).Do(req); err != nil {
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

// Ping 在 Collector 接口的实现方法 Collect() 中
// 让 Exporter 每次获取数据时，都检验一下目标设备通信是否正常
// func (g *GdasClient) Ping() (bool, error) {
// 	// fmt.Println("每次从 Gdas 获取数据时，都会进行测试")
// 	req, err := http.NewRequest("GET", g.Opts.URL+"/configurations", nil)
// 	if err != nil {
// 		return false, err
// 	}
// 	req.SetBasicAuth(g.Opts.Username, g.Opts.password)

// 	resp, err := g.Client.Do(req)
// 	if err != nil {
// 		return false, err
// 	}

// 	resp.Body.Close()

// 	switch {
// 	case resp.StatusCode == http.StatusOK:
// 		return true, nil
// 	case resp.StatusCode == http.StatusUnauthorized:
// 		return false, errors.New("username or password incorrect")
// 	default:
// 		return false, fmt.Errorf("error handling request, http-statuscode: %s", resp.Status)
// 	}
// }
