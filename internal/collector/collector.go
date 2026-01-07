// Package collector 提供市场数据采集功能
package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// MarketData 市场数据结构体
type MarketData struct {
	// AHR999 AHR999指数值
	AHR999 float64 `json:"ahr999"`
	// Timestamp 数据获取时间戳
	Timestamp time.Time `json:"timestamp"`
	// Source 数据来源
	Source string `json:"source"`
}

// Collector 数据采集器
type Collector struct {
	client    *http.Client
	ahr999URL string
	userAgent string
}

// NewCollector 创建新的数据采集器
func NewCollector(ahr999URL string) *Collector {
	return &Collector{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		ahr999URL: ahr999URL,
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
}

// ahr999Response API响应结构
type ahr999Response struct {
	Code int `json:"code"`
	Data struct {
		AHR999 float64 `json:"ahr999"`
	} `json:"data"`
	Message string `json:"message"`
}

// FetchAHR999 获取AHR999指数
func (c *Collector) FetchAHR999() (*MarketData, error) {
	req, err := http.NewRequest("GET", c.ahr999URL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头，模拟浏览器访问
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求返回非200状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	var apiResp ahr999Response
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("API返回错误: %s", apiResp.Message)
	}

	return &MarketData{
		AHR999:    apiResp.Data.AHR999,
		Timestamp: time.Now(),
		Source:    c.ahr999URL,
	}, nil
}

// FetchAHR999FromHTML 从HTML页面解析AHR999指数（备用方案）
// 适用于从Coinglass等网站抓取数据
func (c *Collector) FetchAHR999FromHTML(url string) (*MarketData, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求返回非200状态码: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %w", err)
	}

	// 查找class为"ahr-value"的div元素
	var ahr999Value float64
	var found bool

	doc.Find(".ahr-value").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if _, err := fmt.Sscanf(text, "%f", &ahr999Value); err == nil {
			found = true
		}
	})

	if !found {
		return nil, fmt.Errorf("未找到AHR999数据")
	}

	return &MarketData{
		AHR999:    ahr999Value,
		Timestamp: time.Now(),
		Source:    url,
	}, nil
}
