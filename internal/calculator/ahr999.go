// Package calculator 提供加密货币指标计算功能
package calculator

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// 比特币创世区块日期：2009年1月3日
var genesisDate = time.Date(2009, 1, 3, 0, 0, 0, 0, time.UTC)

// AHR999拟合常数（九神拟合）
const (
	ahr999Coefficient = 5.84
	ahr999Constant    = -17.01
)

// AHR999Calculator AHR999指数计算器
type AHR999Calculator struct {
	client    *http.Client
	baseURL   string
	proxyAddr string
}

// NewAHR999Calculator 创建AHR999计算器
func NewAHR999Calculator() *AHR999Calculator {
	return &AHR999Calculator{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.binance.com",
	}
}

// NewAHR999CalculatorWithProxy 创建带代理的AHR999计算器
func NewAHR999CalculatorWithProxy(proxyAddr string) *AHR999Calculator {
	proxyURL, _ := url.Parse("http://" + proxyAddr)

	return &AHR999Calculator{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		},
		baseURL:   "https://api.binance.com",
		proxyAddr: proxyAddr,
	}
}

// AHR999Result AHR999计算结果
type AHR999Result struct {
	// AHR999 最终指数值
	AHR999 float64
	// CurrentPrice 当前BTC价格
	CurrentPrice float64
	// DCA200Cost 200日定投成本（几何平均）
	DCA200Cost float64
	// ExponentialValue 指数增长估值
	ExponentialValue float64
	// CoinAgeDays 币龄天数
	CoinAgeDays int
	// Timestamp 计算时间
	Timestamp time.Time
}

// Calculate 计算AHR999指数
func (c *AHR999Calculator) Calculate() (*AHR999Result, error) {
	// 1. 获取过去200天的K线数据
	prices, err := c.fetchKlines(200)
	if err != nil {
		return nil, fmt.Errorf("获取K线数据失败: %w", err)
	}

	if len(prices) < 200 {
		return nil, fmt.Errorf("K线数据不足，需要200条，实际获取%d条", len(prices))
	}

	// 当前价格（最新一根K线的收盘价）
	currentPrice := prices[len(prices)-1]

	// 2. 计算200日定投成本（几何平均）
	dca200Cost := calculateGeometricMean(prices)

	// 3. 计算币龄天数
	coinAgeDays := int(time.Since(genesisDate).Hours() / 24)

	// 4. 计算指数增长估值
	exponentialValue := calculateExponentialValue(coinAgeDays)

	// 5. 计算AHR999
	// AHR999 = (当前价格/200日定投成本) × (当前价格/指数增长估值)
	ahr999 := (currentPrice / dca200Cost) * (currentPrice / exponentialValue)

	return &AHR999Result{
		AHR999:           ahr999,
		CurrentPrice:     currentPrice,
		DCA200Cost:       dca200Cost,
		ExponentialValue: exponentialValue,
		CoinAgeDays:      coinAgeDays,
		Timestamp:        time.Now(),
	}, nil
}

// fetchKlines 从Binance获取K线数据
func (c *AHR999Calculator) fetchKlines(limit int) ([]float64, error) {
	// Binance K线API
	// GET /api/v3/klines?symbol=BTCUSDT&interval=1d&limit=200
	apiURL := fmt.Sprintf("%s/api/v3/klines?symbol=BTCUSDT&interval=1d&limit=%d", c.baseURL, limit)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "CryptoSentinel/1.0")

	resp, err := c.client.Do(req)
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

	// Binance K线返回格式: [[开盘时间, 开, 高, 低, 收, 成交量, ...], ...]
	// 我们需要收盘价，索引为4
	var klines [][]interface{}
	if err := json.Unmarshal(body, &klines); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	prices := make([]float64, 0, len(klines))
	for _, kline := range klines {
		if len(kline) < 5 {
			continue
		}
		// 收盘价是字符串格式
		closeStr, ok := kline[4].(string)
		if !ok {
			continue
		}
		closePrice, err := strconv.ParseFloat(closeStr, 64)
		if err != nil {
			continue
		}
		prices = append(prices, closePrice)
	}

	return prices, nil
}

// calculateGeometricMean 计算几何平均数
// 使用对数法避免溢出: exp((ln(P1) + ln(P2) + ... + ln(Pn)) / n)
func calculateGeometricMean(prices []float64) float64 {
	if len(prices) == 0 {
		return 0
	}

	var logSum float64
	for _, price := range prices {
		if price > 0 {
			logSum += math.Log(price)
		}
	}

	return math.Exp(logSum / float64(len(prices)))
}

// calculateExponentialValue 计算指数增长估值
// 公式: 10^(5.84 × log10(币龄天数) - 17.01)
func calculateExponentialValue(coinAgeDays int) float64 {
	if coinAgeDays <= 0 {
		return 0
	}

	// log10(coinAgeDays)
	log10Days := math.Log10(float64(coinAgeDays))

	// 5.84 × log10(币龄天数) - 17.01
	exponent := ahr999Coefficient*log10Days + ahr999Constant

	// 10^exponent
	return math.Pow(10, exponent)
}
