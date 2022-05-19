package collector

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// 这三个常量用于给每个 Metrics 名字添加前缀
const (
	name      = "gdas_exporter"
	Namespace = "gdas"
	//Subsystem(s).
)

// Name 用于给前端页面显示 const 常量中定义的内容
func Name() string {
	return name
}

// GetToken 获取 Gdas 认证所需 Token
func GetToken(opts *GdasOpts) (token string, err error) {
	// 设置 json 格式的 request body
	jsonReqBody := []byte("{\"userName\":\"" + opts.Username + "\",\"passWord\":\"" + opts.Password + "\"}")
	// 设置 URL
	url := fmt.Sprintf("%v/v1/login", opts.URL)
	// 设置 Request 信息
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonReqBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Referer", fmt.Sprintf("%v/gdas", opts.URL))
	req.Header.Add("stime", fmt.Sprintf("%v", time.Now().UnixNano()/1e6))
	// 忽略 TLS 的证书验证
	ts := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 发送 Request 并获取 Response
	resp, err := (&http.Client{Transport: ts}).Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{
			"reson": "发起 HTTP 请求异常",
			"code":  resp.StatusCode,
		}).Errorf("GetToken Error")
		return
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
	token, err = jsonRespBody.Get("token").String()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"reson": "获取响应体中的数据失败",
		}).Errorf("GetToken Error")
		return
	}
	logrus.WithFields(logrus.Fields{
		"Token": token,
	}).Debugf("Get Token Successed!")
	return
}

// ######## 从此处开始到文件结尾，都是关于配置连接 Gdas 的代码 ########

// GdasClient 连接 Gdas 所需信息
type GdasClient struct {
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
	return &GdasClient{
		Opts:  opts,
		Token: token,
		Client: &http.Client{
			Timeout:   opts.Timeout,
			Transport: transport,
		},
	}
}

// Request 建立与 Gdas 的连接，并返回 Response Body
func (c *GdasClient) Request(method string, endpoint string, reqBody io.Reader) (body []byte, err error) {
	// 根据认证信息及 endpoint 参数，创建与 Gdas 的连接，并返回 Body 给每个 Metric 采集器
	url := c.Opts.URL + endpoint
	logrus.WithFields(logrus.Fields{
		"url":    url,
		"method": method,
	}).Debugf("抓取指标时的请求URL")

	randString, signatureSha := generateSign(c.Token)

	// 创建一个新的 Request
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Opts.Username, c.Opts.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", c.Token)
	req.Header.Set("stime", fmt.Sprintf("%v", time.Now().UnixNano()/1e6))
	req.Header.Set("nonce", randString)
	req.Header.Set("signature", signatureSha)
	req.Header.Set("Referer", fmt.Sprintf("%v/gdas", c.Opts.URL))

	// 根据新建立的 Request，发起请求，并获取 Response
	resp, err := c.Client.Do(req)
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
	logrus.WithFields(logrus.Fields{
		"code": resp.StatusCode,
		"body": string(body),
	}).Tracef("每次请求的响应体以及响应状态码")
	return body, nil
}

// Ping 在 Scraper 接口的实现方法 scrape() 中调用。
// 让 Exporter 每次获取数据时，都检验一下目标设备通信是否正常
func (c *GdasClient) Ping() (b bool, err error) {
	// 判断 TOKEN 是否可用
	url := c.Opts.URL + "/v1/nodeList"
	method := "GET"
	logrus.WithFields(logrus.Fields{
		"url":    url,
		"method": method,
	}).Debugf("每次从 Gdas 并发抓取指标之前，先检查一下目标状态")

	randString, signatureSha := generateSign(c.Token)

	// 创建一个新的 Request
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", c.Token)
	req.Header.Set("stime", fmt.Sprintf("%v", time.Now().UnixNano()/1e6))
	req.Header.Set("nonce", randString)
	req.Header.Set("signature", signatureSha)
	req.Header.Set("Referer", fmt.Sprintf("%v/gdas", c.Opts.URL))

	// 根据新建立的 Request，发起请求，并获取 Response
	resp, err := c.Client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
		return true, nil
	case resp.StatusCode == http.StatusUnauthorized:
		fmt.Printf("认证检查失败，状态码为：%v,尝试重新获取 Token\n", resp.Status)
		if c.Token, err = GetToken(c.Opts); err != nil {
			return false, err
		}
		return true, nil
	default:
		fmt.Println("检查失败，状态码为：", resp.Status)
		c.Token, err = GetToken(c.Opts)
		if err != nil {
			return false, fmt.Errorf("error handling request, http-statuscode: %s,http-ResponseBody：%s", resp.Status, resp.Body)
		}
		return true, nil
	}
}

func (c *GdasClient) GetConcurrency() int {
	return c.Opts.Concurrency
}

// GdasOpts 登录 Gdas 所需属性
type GdasOpts struct {
	URL         string
	Username    string
	Password    string
	Concurrency int
	// 这俩是关于 http.Client 的选项
	Timeout  time.Duration
	Insecure bool
}

// AddFlag use after set Opts
func (o *GdasOpts) AddFlag() {
	pflag.StringVar(&o.URL, "gdas-server", "https://172.38.30.193:8003", "HTTP API address of a Gdas server or agent. (prefix with https:// to connect over HTTPS)")
	pflag.StringVar(&o.Username, "gdas-user", "system", "gdas username")
	pflag.StringVar(&o.Password, "gdas-pass", "", "gdas password")
	pflag.IntVar(&o.Concurrency, "concurrent", 10, "Number of concurrent requests during collection.")
	pflag.DurationVar(&o.Timeout, "time-out", time.Millisecond*1600, "Timeout on HTTP requests to the Gads API.")
	pflag.BoolVar(&o.Insecure, "insecure", true, "Disable TLS host verification.")
}

// 生成签名所需数据
func generateSign(token string) (string, string) {

	// 毫秒时间戳
	stime := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	// 随机数
	randString := rand.Intn(100000)
	// 随机数倒序
	stringRand := []rune(strconv.Itoa(randString))
	for from, to := 0, len(stringRand)-1; from < to; from, to = from+1, to-1 {
		stringRand[from], stringRand[to] = stringRand[to], stringRand[from]
	}
	// 签名
	signature := stime + strconv.Itoa(randString) + token + string(stringRand)
	h := sha256.New()
	h.Write([]byte(signature))                     // 需要加密的字符串为
	signatureSha := hex.EncodeToString(h.Sum(nil)) // 输出加密结果

	return strconv.Itoa(randString), signatureSha
}
