package collector

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// 这三个常量用于给每个 Metrics 名字添加前缀
const (
	name      = "e37_exporter"
	Namespace = "e37"
	//Subsystem(s).
	// exporter = "exporter"
)

// Name 用于给前端页面显示 const 常量中定义的内容
func Name() string {
	return name
}

// GetToken 获取 E37 认证所需 Token
func GetToken(opts *E37Opts) (token string, err error) {
	// 设置 json 格式的 request body
	jsonReqBody := []byte("{\"username\":\"" + opts.Username + "\",\"password\":\"" + opts.password + "\"}")
	// 设置 URL
	url := fmt.Sprintf("%v/api/auth", opts.URL)
	// 设置 Request 信息
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonReqBody))
	// req.Header.Add("Accept", "")
	// req.Header.Add("Content-Type", "")
	req.Header.Set("User-Agent", "apifox/1.0.0 (https://www.apifox.cn)")
	req.Header.Set("Content-Type", "application/json")
	// 忽略 TLS 的证书验证
	ts := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 发送 Request 并获取 Response
	resp, err := (&http.Client{Transport: ts}).Do(req)
	if err != nil || resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GetToken Error: %v\nReason:%v", resp.StatusCode, string(respBody))
	}
	defer resp.Body.Close()

	// 处理 Response Body,并获取 Token
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	jsonRespBody, err := simplejson.NewJson(respBody)
	if err != nil {
		return
	}
	logrus.Debugf("Get Token Status:\nResponseStatusCode：%v\nResponseBody：%v\n", resp.StatusCode, string(respBody))
	token, err = jsonRespBody.Get("token").String()
	if err != nil {
		return "", fmt.Errorf("GetToken Error：%v", err)
	}
	logrus.Debugf("Get Token Successed!Token is:%v ", token)
	return
}

// ######## 从此处开始到文件结尾，都是关于配置连接 E37 的代码 ########

// E37Client 连接 E37 所需信息
type E37Client struct {
	Client *http.Client
	Token  string
	Opts   *E37Opts
}

// NewE37Client 实例化 E37 客户端
func NewE37Client(opts *E37Opts) *E37Client {
	uri := opts.URL
	if !strings.Contains(uri, "://") {
		uri = "http://" + uri
	}
	u, err := url.Parse(uri)
	if err != nil {
		panic(fmt.Sprintf("invalid E37 URL: %s", err))
	}
	if u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		panic(fmt.Sprintf("invalid E37 URL: %s", uri))
	}

	// ######## 配置 http.Client 的信息 ########
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		panic(err)
	}
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

	//
	token, err := GetToken(opts)
	if err != nil {
		panic(err)
	}
	return &E37Client{
		Opts:  opts,
		Token: token,
		Client: &http.Client{
			Timeout:   opts.Timeout,
			Transport: transport,
		},
	}
}

// Request 建立与 E37 的连接，并返回 Response Body
func (g *E37Client) Request(method string, endpoint string, reqBody io.Reader) (body []byte, err error) {
	// 根据认证信息及 endpoint 参数，创建与 E37 的连接，并返回 Body 给每个 Metric 采集器
	url := g.Opts.URL + endpoint
	logrus.Debugf("request url is: %s", url)

	// 创建一个新的 Request
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", g.Token))

	// 根据新建立的 Request，发起请求，并获取 Response
	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error handling request for %s http-statuscode: %s", endpoint, resp.Status)
	}

	// 处理 Response Body
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// logrus.Debugf("Response Status:\nResponseStatusCode：%v\nResponseBody：%v\n", resp.StatusCode, string(body))
	return body, nil
}

// 验证 Token 时所用的请求体
type token struct {
	Token string `json:"token"`
}

// Ping 在 Scraper 接口的实现方法 scrape() 中调用。
// 让 Exporter 每次获取数据时，都检验一下目标设备通信是否正常
func (g *E37Client) Ping() (b bool, err error) {
	logrus.Debugf("每次从 E37 并发抓取指标之前，先检查一下目标状态")

	// 判断 TOKEN 是否可用
	url := g.Opts.URL + "/api/auth/check"
	logrus.Debugf("Ping Request url %s", url)

	t := token{
		Token: g.Token,
	}
	jsonReqBody, err := json.Marshal(t)
	if err != nil {
		return false, err
	}
	// jsonReqBody := []byte("{\"token\":\"" + g.Token + "\"}")

	// 创建一个新的 Request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonReqBody))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	// 根据新建立的 Request，发起请求，并获取 Response
	resp, err := g.Client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	jsonRespBody, err := simplejson.NewJson(respBody)
	if err != nil {
		return false, err
	}
	// 若响应体没有 username 字段，则重新获取 Token
	_, err = jsonRespBody.Get("username").String()
	if err != nil {
		logrus.Errorf("Token 检查失败，状态码为：%v,尝试重新获取 Token\n", resp.Status)
		g.Token, err = GetToken(g.Opts)
		if err != nil {
			return false, fmt.Errorf("重新获取 Token 失败，响应吗：%v，响应体：%v", resp.Status, resp.Body)
		}
		return true, nil
	} else {
		return true, nil
	}
}

// E37Opts 登录 E37 所需属性
type E37Opts struct {
	URL      string
	Username string
	password string
	// 这俩是关于 http.Client 的选项
	Timeout  time.Duration
	Insecure bool
}

// AddFlag use after set Opts
func (o *E37Opts) AddFlag() {
	pflag.StringVar(&o.URL, "e37-server", "https://172.38.30.2:8443", "HTTP API address of a E37 server or agent. (prefix with https:// to connect over HTTPS)")
	pflag.StringVar(&o.Username, "e37-user", "admin", "e37 username")
	pflag.StringVar(&o.password, "e37-pass", "admin", "e37 password")
	pflag.DurationVar(&o.Timeout, "time-out", time.Millisecond*1600, "Timeout on HTTP requests to the Gads API.")
	pflag.BoolVar(&o.Insecure, "insecure", true, "Disable TLS host verification.")
}