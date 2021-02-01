package collector

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
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
	name      = "xsky_exporter"
	namespace = "xsky"
	//Subsystem(s).
	exporter = "exporter"
)

// Name 用于给前端页面显示 const 常量中定义的内容
func Name() string {
	return name
}

// GetToken 获取 Xsky 认证所需 Token
func GetToken(opts *XskyOpts) (token string, err error) {
	// 设置 json 格式的 request body
	jsonReqBody := []byte("{\"auth\":{\"name\":\"" + opts.Username + "\",\"password\":\"" + opts.password + "\"}}")
	// 设置 URL
	url := fmt.Sprintf("%v/api/v1/auth/tokens:login", opts.URL)
	// 设置 Request 信息
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonReqBody))
	req.Header.Add("Content-Type", "application/json")
	// 忽略 TLS 的证书验证
	ts := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// 发送 Request 并获取 Response
	resp, err := (&http.Client{Transport: ts}).Do(req)
	if err != nil || resp.StatusCode != http.StatusCreated {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("GetToken Error: %v\nResonse:%v", resp.StatusCode, string(respBody))
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
	logrus.Debugf("Get Token Status:\nResponseStatusCode：%v\nResponseBody：%v\n", resp.StatusCode, string(respBody))
	if token, err = jsonRespBody.Get("token").Get("uuid").String(); err != nil {
		return "", fmt.Errorf("GetToken Error：%v", err)
	}
	logrus.Debugf("Get Token Successed!Token is:%v ", token)
	return
}

// ######## 从此处开始到文件结尾，都是关于配置连接 Xsky 的代码 ########

// XskyClient 连接 Xsky 所需信息。实现了 CommonClient 接口
type XskyClient struct {
	Client *http.Client
	Token  string
	Opts   *XskyOpts
}

// NewXsykClient 实例化 Xsky 客户端
func NewXsykClient(opts *XskyOpts) *XskyClient {
	uri := opts.URL
	if !strings.Contains(uri, "://") {
		uri = "http://" + uri
	}
	u, err := url.Parse(uri)
	if err != nil {
		panic(fmt.Sprintf("invalid Xsky URL: %s", err))
	}
	if u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		panic(fmt.Sprintf("invalid Xsky URL: %s", uri))
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

	//
	token, err := GetToken(opts)
	if err != nil {
		panic(err)
	}

	return &XskyClient{
		Opts:  opts,
		Token: token,
		Client: &http.Client{
			Timeout:   opts.Timeout,
			Transport: transport,
		},
	}
}

// Request 建立与 Xsky 的连接，并返回 Response Body
func (x *XskyClient) Request(method string, endpoint string, reqBody io.Reader) (body []byte, err error) {
	// 根据认证信息及 endpoint 参数，创建与 Xsky 的连接，并返回 Body 给每个 Metric 采集器
	url := x.Opts.URL + endpoint
	logrus.Debugf("request url %s", url)

	// 创建一个新的 Request
	// req, err := http.NewRequest("GET", url, nil)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(x.Opts.Username, x.Opts.password)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Xms-Auth-Token", x.Token)

	// 根据新建立的 Request，发起请求，并获取 Response
	resp, err := x.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error handling request for %s http-statuscode: %s", endpoint, resp.Status)
	}

	// 处理 Response Body
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// logrus.Debugf("Response Status:\nResponseStatusCode：%v\nResponseBody：%v\n", resp.StatusCode, string(body))
	return body, nil
}

// Ping 在 Scraper 接口的实现方法 scrape() 中调用。
// 让 Exporter 每次获取数据时，都检验一下目标设备通信是否正常
func (x *XskyClient) Ping() (b bool, err error) {
	logrus.Debugf("每次从 Xsky 并发抓取指标之前，先检查一下目标状态")
	// 判断是否有 Token
	if x.Token == "" {
		logrus.Debugf("Token 为空，开始尝试获取 Token")
		x.Token, err = GetToken(x.Opts)
		if err == nil {
			return true, nil
		}
		return false, err
	}
	logrus.Debugf("Xsky Token 为: %s", x.Token)

	// TODO 还需要添加一个认证，当 Token 失效时，也需要重新获取 Token，可以直接
	logrus.Debugf("Ping Request url %s", x.Opts.URL+"/health")
	req, err := http.NewRequest("GET", x.Opts.URL+"/health", nil)
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
	pflag.StringVar(&o.URL, "xsky-server", "http://10.20.5.98:8056", "HTTP API address of a Xsky server or agent. (prefix with https:// to connect over HTTPS)")
	pflag.StringVar(&o.Username, "xsky-user", "admin", "xsky username")
	pflag.StringVar(&o.password, "xsky-pass", "", "xsky password")
	pflag.DurationVar(&o.Timeout, "time-out", time.Millisecond*1600, "Timeout on HTTP requests to the Xsky API.")
	pflag.BoolVar(&o.Insecure, "insecure", true, "Disable TLS host verification.")
}
