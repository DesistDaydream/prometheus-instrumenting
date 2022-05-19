package collector

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/rand"
	"strconv"

	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// 这三个常量用于给每个 Metrics 名字添加前缀
const (
	name      = "console_agent_exporter"
	Namespace = "console_agent"
	exporter  = "exporter"
)

// Name 用于给前端页面显示 const 常量中定义的内容
func Name() string {
	return name
}

// ConsoleAgentRespBody 控制台返回的响应体
type ConsoleAgentRespBody struct {
	TraceId   string      `json:"traceId"`
	Timestamp int64       `json:"timestamp"`
	Code      string      `json:"code"` // 控制台响应码
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"` // Gdas返回给控制台的数据都在 Data 中
}

// ConsoleAgentClient 连接 ConsoleAgent 所需信息
type ConsoleAgentClient struct {
	Client *http.Client
	Opts   *ConsoleAgentOpts
}

// NewConsoleAgentClient 实例化 ConsoleAgent 客户端
func NewConsoleAgentClient(opts *ConsoleAgentOpts) *ConsoleAgentClient {
	uri := opts.URL
	if !strings.Contains(uri, "://") {
		uri = "http://" + uri
	}
	u, err := url.Parse(uri)
	if err != nil {
		panic(fmt.Sprintf("invalid ConsoleAgent URL: %s", err))
	}
	if u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		panic(fmt.Sprintf("invalid ConsoleAgent URL: %s", uri))
	}

	return &ConsoleAgentClient{
		Opts: opts,
		Client: &http.Client{
			Timeout: opts.Timeout,
		},
	}
}

// Request 建立与 ConsoleAgent 的连接，并返回 Response Body
func (c *ConsoleAgentClient) Request(method string, endpoint string, reqBody io.Reader) (data []byte, err error) {
	// 根据认证信息及 endpoint 参数，创建与 ConsoleAgent 的连接，并返回 Body 给每个 Metric 采集器
	urls := c.Opts.URL + endpoint

	// 创建一个新的 Request
	req, err := http.NewRequest(method, urls, reqBody)
	if err != nil {
		return nil, err
	}

	// 为 HTTP Request 设置 Header 参数
	// 由于要并发建立多个请求，所有请求头里的时间戳会变化，所以每个请求都要分配一块内存空间来存放数据。
	// 如果不是因为这个原因，可以直接把请求头的信息，直接写道 ConsoleAgentOpts 结构体中。
	r := c.NewReqHeaderValues()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("appkey", r.Appkey)
	req.Header.Set("stime", r.Stimestamp)
	req.Header.Set("nonce", r.Nonce)
	req.Header.Set("signature", r.Signature)

	logrus.Debugf("Request Method and URL is: %s %s\n", req.Method, req.URL)

	// 根据新建立的 Request，发起请求，并获取 Response
	var resp *http.Response
	if resp, err = c.Client.Do(req); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error handling request for %s http-statuscode: %s", endpoint, resp.Status)
	}

	var consoleAgentRespBody ConsoleAgentRespBody
	// 处理 Response Body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// logrus.Debugf("当前响应体为：", string(respBody))
	logrus.Debugf("本次请求的响应码为：%v", resp.StatusCode)

	//提取出控制台返回的响应体中关于 Gdas 的数据，并交给各个 Scrape 使用
	err = json.Unmarshal(respBody, &consoleAgentRespBody)
	if err != nil {
		return nil, err
	}
	data, _ = json.Marshal(consoleAgentRespBody.Data)
	return data, nil
}

// Ping 在 Scraper 接口的实现方法 scrape() 中调用。
// 让 Exporter 每次获取数据时，通过控制台验证和ConsoleAgent的连接
func (c *ConsoleAgentClient) Ping() (b bool, err error) {
	logrus.Debugf("每次从 ConsoleAgent 并发抓取指标之前，先检查一下目标状态")
	req, err := http.NewRequest("GET", c.Opts.URL+"/api/actuator/health", nil)
	if err != nil {
		return false, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return false, err
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, nil
	}
}

// reqHeaderValues 请求头中应该具有的字段
type reqHeaderValues struct {
	Appkey     string
	Stimestamp string
	Nonce      string
	Signature  string
}

// NewReqHeaderValues 生成请求头的内容
func (c *ConsoleAgentClient) NewReqHeaderValues() *reqHeaderValues {
	// 接入渠道标识
	appkey := "wo-obs"
	secretKey := c.Opts.SecretKey
	// 毫秒时间戳
	stimestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	// 随机字符串
	const char = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.NewSource(time.Now().UnixNano()) // 产生随机种子
	var s bytes.Buffer
	for i := 0; i < 20; i++ {
		s.WriteByte(char[rand.Int63()%int64(len(char))])
	}
	nonce := s.String()
	// 随机字符串的逆序
	var bytes []byte = []byte(nonce)
	for i := 0; i < len(nonce)/2; i++ {
		tmp := bytes[len(nonce)-i-1]
		bytes[len(nonce)-i-1] = bytes[i]
		bytes[i] = tmp
	}
	nonceReverse := string(bytes)
	// 签名 secretKey+nonce+stime+nonce的倒序拼接再SHA256加密
	signOriginal := secretKey + nonce + stimestamp + nonceReverse
	// SHA256加密
	h := sha256.New()
	h.Write([]byte(signOriginal))
	signEncrypt := h.Sum(nil)
	signature := hex.EncodeToString(signEncrypt)

	return &reqHeaderValues{
		Appkey:     appkey,
		Stimestamp: stimestamp,
		Nonce:      nonce,
		Signature:  signature,
	}
}

func (c *ConsoleAgentClient) GetConcurrency() int {
	return c.Opts.Concurrency
}

// ConsoleAgentOpts 登录 ConsoleAgent 所需属性
type ConsoleAgentOpts struct {
	URL         string
	Concurrency int
	//http.Client的选项
	Timeout   time.Duration
	Insecure  bool
	SecretKey string
}

// AddFlag use after set Opts
func (o *ConsoleAgentOpts) AddFlag() {
	pflag.StringVar(&o.URL, "console-agent-server", "http://172.38.40.210:9097", "HTTP API address of a harbor server or agent. (prefix with https:// to connect over HTTPS)")
	pflag.IntVar(&o.Concurrency, "concurrency", 10, "Number of concurrency requests during collection.")
	pflag.DurationVar(&o.Timeout, "time-out", time.Millisecond*60000, "Timeout on HTTP requests to the Gdas-Proxy.")
	pflag.BoolVar(&o.Insecure, "insecure", true, "Disable TLS host verification.")
	pflag.StringVar(&o.SecretKey, "secret-key", "obs123456", "Set Http Request Header SecretKey.")
}
