package collector

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// 这三个常量用于给每个 Metrics 名字添加前缀
const (
	name      = "gdas_exporter"
	namespace = "gdas"
	//Subsystem(s).
	exporter = "exporter"
)

// Name 用于给前端页面显示 const 常量中定义的内容
func Name() string {
	return name
}

// GdasClient 连接 Gdas 所需信息
type GdasClient struct {
	req    *http.Request
	resp   *http.Response
	Client *http.Client
	Token  string
	Opts   *GdasOpts
}

// NewGdasClient 实例化 Gdas 客户端
func NewGdasClient(opts *GdasOpts) *GdasClient {
	uri := opts.URL
	if !strings.Contains(uri, "://") {
		uri = "http://" + uri
	}
	u, err := url.Parse(uri)
	if err != nil {
		panic(fmt.Sprintf("invalid Gdas URL: %s", err))
	}
	if u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		panic(fmt.Sprintf("invalid Gdas URL: %s", uri))
	}

	// ######## 配置 http.Client 的信息 ########
	rootCAs, err := x509.SystemCertPool()
	// if err != nil {
	// 	return nil, err
	// }
	// 初始化 TLS 相关配置信息
	tlsClientConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    rootCAs,
	}
	// 可以通过命令行选项配置 TLS 的 InsecureSkipVerify
	// 这个配置决定是否跳过 https 协议的验证过程，就是 curl 加不加 -k 选项。默认跳过
	if opts.Insecure {
		tlsClientConfig.InsecureSkipVerify = true
	}
	transport := &http.Transport{
		TLSClientConfig: tlsClientConfig,
	}
	// ######## 配置 http.Client 的信息结束 ########

	token, _ := GetToken(opts)
	return &GdasClient{
		Opts:  opts,
		Token: token,
		Client: &http.Client{
			Timeout:   opts.Timeout,
			Transport: transport,
		},
	}
}

// GetToken 获取 Gdas 认证所需 Token
func GetToken(opts *GdasOpts) (token string, err error) {
	// 设置 json 格式的 request body
	jsonReqBody := []byte("{\"userName\":\"" + opts.Username + "\",\"passWord\":\"" + opts.password + "\"}")
	// 设置 URL
	url := fmt.Sprintf("%v/v1/login", opts.URL)
	// 设置 Request 信息
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonReqBody))
	req.Header.Add("referer", fmt.Sprintf("%v/v1/login", opts.URL))
	req.Header.Add("Content-Type", "application/json")
	// 忽略 TLS 的证书验证
	ts := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 发送 Request 并获取 Response
	resp, err := (&http.Client{Transport: ts}).Do(req)
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
	token, _ = jsonRespBody.Get("token").String()
	fmt.Println("成功获取 Token！ ", token)
	return
}

// Request 建立与 Gdas 的连接，并返回 Response Body
func (g *GdasClient) Request(endpoint string) (body []byte, err error) {
	// 获取 Gdas 认证所需 Token
	if err = g.RequestCheck(endpoint); err != nil {
		fmt.Println(err)
		GetToken(g.Opts)
	}
	fmt.Println("Gdas Token 为：", g.Token)

	// 根据认证信息及 endpoint 参数，创建与 Gdas 的连接，并返回 Body 给每个 Metric 采集器
	url := g.Opts.URL + endpoint
	logrus.Debugf("request url %s", url)

	// 创建一个新的 Request
	if g.req, err = http.NewRequest("GET", url, nil); err != nil {
		return nil, err
	}
	g.req.SetBasicAuth(g.Opts.Username, g.Opts.password)
	g.req.Header.Set("Content-Type", "application/json; charset=utf-8")
	g.req.Header.Set("token", g.Token)

	// 根据新建立的 Request，发起请求，并获取 Response
	if g.resp, err = g.Client.Do(g.req); err != nil {
		return nil, err
	}
	defer g.resp.Body.Close()

	if g.resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error handling request for %s http-statuscode: %s", endpoint, g.resp.Status)
	}

	// 处理 Response Body
	if body, err = ioutil.ReadAll(g.resp.Body); err != nil {
		return nil, err
	}

	return body, nil
}

// RequestCheck 检查当前请求的认证信息是否正确
func (g *GdasClient) RequestCheck(endpoint string) (err error) {
	// 判断是否有 TOKEN
	if g.Token == "" {
		return fmt.Errorf("处理请求出错：没有 Token")
	}

	// 判断 TOKEN 是否可用
	url := g.Opts.URL + endpoint
	logrus.Debugf("request url %s", url)

	// 创建一个新的 Request
	g.req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	g.req.Header.Set("Content-Type", "application/json; charset=utf-8")
	g.req.Header.Set("token", g.Token)
	g.req.Header.Add("referer", fmt.Sprintf("%v/v1/login", g.Opts.URL))

	// 根据新建立的 Request，发起请求，并获取 Response
	if g.resp, err = g.Client.Do(g.req); err != nil {
		return err
	}
	defer g.resp.Body.Close()

	if g.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error handling request for %s http-statuscode: %s，Token 不可用", endpoint, g.resp.Status)
	}

	return nil
}

// Ping 在 Scraper 接口的实现方法 scrape() 中调用。
// 让 Exporter 每次获取数据时，都检验一下目标设备通信是否正常
func (g *GdasClient) Ping() (b bool, err error) {
	if g.req, err = http.NewRequest("GET", g.Opts.URL+"/待求证健康检查接口", nil); err != nil {
		return false, err
	}

	if g.resp, err = g.Client.Do(g.req); err != nil {
		return false, err
	}

	g.resp.Body.Close()

	switch {
	case g.resp.StatusCode == http.StatusOK:
		return true, nil
	case g.resp.StatusCode == http.StatusUnauthorized:
		return false, errors.New("username or password incorrect")
	default:
		return false, fmt.Errorf("error handling request, http-statuscode: %s", g.resp.Status)
	}
}

// GdasOpts 登录 Gdas 所需属性
type GdasOpts struct {
	URL      string
	Username string
	password string
	// 这俩是关于 http.Client 的选项
	Timeout  time.Duration
	Insecure bool
}

// AddFlag use after set Opts
func (o *GdasOpts) AddFlag() {
	pflag.StringVar(&o.URL, "gdas-server", "https://172.38.30.192:8003", "HTTP API address of a harbor server or agent. (prefix with https:// to connect over HTTPS)")
	pflag.StringVar(&o.Username, "gdas-user", "system", "gdas username")
	pflag.StringVar(&o.password, "gdas-pass", "d153850931040e5c81e1c7508ded25f5f0ae76cb57dc1997bc343b878946ba23", "gdas password")
	pflag.DurationVar(&o.Timeout, "time-out", time.Millisecond*1600, "Timeout on HTTP requests to the harbor API.")
	pflag.BoolVar(&o.Insecure, "insecure", true, "Disable TLS host verification.")
}
