// Package fetcher 提供远程指标数据获取功能
package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// MVRVFetcher MVRV数据获取器
// 数据来源：CoinMetrics Community API (免费)
type MVRVFetcher struct {
	client  *http.Client
	baseURL string
}

// NewMVRVFetcher 创建MVRV获取器
func NewMVRVFetcher() *MVRVFetcher {
	return &MVRVFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://community-api.coinmetrics.io/v4",
	}
}

// NewMVRVFetcherWithProxy 创建带代理的MVRV获取器
func NewMVRVFetcherWithProxy(proxyAddr string) *MVRVFetcher {
	proxyURL, _ := url.Parse("http://" + proxyAddr)

	return &MVRVFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		},
		baseURL: "https://community-api.coinmetrics.io/v4",
	}
}

// MVRVResult MVRV获取结果
type MVRVResult struct {
	// MVRV 市值与已实现价值比率
	// > 3.5 通常表示高估
	// < 1 通常表示低估
	MVRV float64
	// ZScore MVRV Z-Score（标准化后的值）
	// 需要历史数据计算，暂用简化估算
	ZScore float64
	// Timestamp 数据时间
	Timestamp time.Time
	// Source 数据来源
	Source string
}

// coinMetricsResponse CoinMetrics API响应结构
type coinMetricsResponse struct {
	Data []struct {
		Asset      string `json:"asset"`
		Time       string `json:"time"`
		CapMVRVCur string `json:"CapMVRVCur"`
	} `json:"data"`
}

// Fetch 获取最新MVRV数据
func (f *MVRVFetcher) Fetch() (*MVRVResult, error) {
	// CoinMetrics API: 获取BTC的MVRV指标
	apiURL := fmt.Sprintf("%s/timeseries/asset-metrics?assets=btc&metrics=CapMVRVCur&frequency=1d&page_size=1&api_key=community", f.baseURL)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "CryptoSentinel/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回非200状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var apiResp coinMetricsResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("API返回数据为空")
	}

	// 解析MVRV值
	mvrvStr := apiResp.Data[0].CapMVRVCur
	mvrv, err := strconv.ParseFloat(mvrvStr, 64)
	if err != nil {
		return nil, fmt.Errorf("解析MVRV值失败: %w", err)
	}

	// 解析时间
	dataTime, _ := time.Parse(time.RFC3339Nano, apiResp.Data[0].Time)

	// 简化Z-Score估算
	// 历史均值约 1.5，标准差约 1.2
	// Z-Score = (MVRV - 均值) / 标准差
	historicalMean := 1.5
	historicalStd := 1.2
	zScore := (mvrv - historicalMean) / historicalStd

	return &MVRVResult{
		MVRV:      mvrv,
		ZScore:    zScore,
		Timestamp: dataTime,
		Source:    "CoinMetrics Community API",
	}, nil
}

// FetchWithHistory 获取MVRV并使用真实历史数据计算Z-Score
func (f *MVRVFetcher) FetchWithHistory(days int) (*MVRVResult, error) {
	// 获取历史数据计算真实Z-Score
	apiURL := fmt.Sprintf("%s/timeseries/asset-metrics?assets=btc&metrics=CapMVRVCur&frequency=1d&page_size=%d&api_key=community", f.baseURL, days)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "CryptoSentinel/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回非200状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var apiResp coinMetricsResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("API返回数据为空")
	}

	// 收集所有MVRV值
	mvrvValues := make([]float64, 0, len(apiResp.Data))
	for _, d := range apiResp.Data {
		if v, err := strconv.ParseFloat(d.CapMVRVCur, 64); err == nil {
			mvrvValues = append(mvrvValues, v)
		}
	}

	if len(mvrvValues) == 0 {
		return nil, fmt.Errorf("无有效MVRV数据")
	}

	// 最新MVRV（数据按时间倒序，第一条是最新的）
	currentMVRV := mvrvValues[0]

	// 计算均值
	var sum float64
	for _, v := range mvrvValues {
		sum += v
	}
	mean := sum / float64(len(mvrvValues))

	// 计算标准差
	var varianceSum float64
	for _, v := range mvrvValues {
		varianceSum += (v - mean) * (v - mean)
	}
	std := 0.0
	if len(mvrvValues) > 1 {
		std = varianceSum / float64(len(mvrvValues)-1)
		if std > 0 {
			std = sqrt(std)
		}
	}

	// 计算Z-Score
	zScore := 0.0
	if std > 0 {
		zScore = (currentMVRV - mean) / std
	}

	// 解析时间
	dataTime, _ := time.Parse(time.RFC3339Nano, apiResp.Data[0].Time)

	return &MVRVResult{
		MVRV:      currentMVRV,
		ZScore:    zScore,
		Timestamp: dataTime,
		Source:    fmt.Sprintf("CoinMetrics Community API (%d天历史)", len(mvrvValues)),
	}, nil
}

// sqrt 简单平方根实现
func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 100; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}
