package collector

import (
	"io/ioutil"
	"net/http"

	"github.com/spf13/pflag"
)

// HarborConnInfo 连接 Harbor 所需信息
type HarborConnInfo struct {
	BaseURL  string
	Username string
	Password string
}

// HC 存储 Harbor 的连接信息
var HC HarborConnInfo

// HarborConnFlags 通过命令行设定连接 Harbor 的信息
func (h *HarborConnInfo) HarborConnFlags() {
	pflag.StringVar(&h.BaseURL, "harbor-baseurl", "http://172.19.42.218/api/v2.0", "Harbor URL")
	pflag.StringVar(&h.Username, "harbor-user", "admin", "Harbor 用户名")
	pflag.StringVar(&h.Password, "harbor-pass", "Harbor12345", "Harbor 密码")
}

// HarborConn 根据指定的 endpoint 连接 Harbor API 并返回 Response Body 以供各采集器处理获取想要的数据
func (h *HarborConnInfo) HarborConn(endpoint string) (body []byte, err error) {
	url := h.BaseURL + endpoint
	// 构建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(h.Username, h.Password)
	req.Header.Set("Content-Type", "application/json")

	// 获取 Response
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 处理 Response Body
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}
	return
}

func harbor() {
	var h HarborConnInfo
	// 加载关于 Harbor 相关的 Flags
	h.HarborConnFlags()
	pflag.Parse()
}

// Conn 设置连接 Harbor 的信息
func Conn() {
	// 加载关于 Harbor 相关的 Flags
	HC.HarborConnFlags()
	pflag.Parse()
}
