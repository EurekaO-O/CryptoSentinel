// Package collector 提供市场数据采集功能
package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"CryptoSentinel/internal/model"
)

// CollectorV2 升级版数据采集器，支持多指标采集
type CollectorV2 struct {
	client    *http.Client
	ahr999URL string
	userAgent string
}

// NewCollectorV2 创建升级版数据采集器
func NewCollectorV2(ahr999URL string) *CollectorV2 {
	return &CollectorV2{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		ahr999URL: ahr999URL,
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
}

// NewCollectorV2WithProxy 创建带代理的数据采集器
func NewCollectorV2WithProxy(ahr999URL, proxyAddr string) *CollectorV2 {
	proxyURL, _ := url.Parse("http://" + proxyAddr)

	return &CollectorV2{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		},
		ahr999URL: ahr999URL,
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
}

// FetchAllIndicators 获取所有市场指标
// leverage: 当前账户杠杆率（由外部传入）
func (c *CollectorV2) FetchAllIndicators(leverage float64) (*model.MarketIndicators, error) {
	indicators := &model.MarketIndicators{
		Timestamp:       time.Now(),
		AccountLeverage: leverage,
		Source:          c.ahr999URL,
	}

	// 1. 获取AHR999指数
	if err := c.fetchAHR999(indicators); err != nil {
		return nil, fmt.Errorf("获取AHR999失败: %w", err)
	}

	// 2. 获取BTC和ETH价格（Mock或从其他API获取）
	if err := c.fetchPrices(indicators); err != nil {
		// 价格获取失败不阻断流程，使用Mock数据
		indicators.CurrentPriceBTC = 0
		indicators.CurrentPriceETH = 0
	}

	// 3. 获取MVRV-Z Score（Mock数据，预留接口）
	indicators.MVRVZScore = c.fetchMVRVZScore()

	// 4. 计算MA乘数状态（Mock数据，预留接口）
	indicators.MaMultiplierState = c.calculateMaMultiplierState(indicators.CurrentPriceBTC)

	// 5. 获取Pi周期状态（Mock数据，预留接口）
	indicators.PiCycleCross = c.fetchPiCycleStatus()

	// 6. 获取ETH回归带位置（Mock数据，预留接口）
	indicators.EthRegressionState = c.calculateEthRegressionState(indicators.CurrentPriceETH)

	return indicators, nil
}

// fetchAHR999 获取AHR999指数
func (c *CollectorV2) fetchAHR999(indicators *model.MarketIndicators) error {
	req, err := http.NewRequest("GET", c.ahr999URL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求返回非200状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应体失败: %w", err)
	}

	var apiResp struct {
		Code int `json:"code"`
		Data struct {
			AHR999 float64 `json:"ahr999"`
		} `json:"data"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("解析JSON失败: %w", err)
	}

	if apiResp.Code != 0 {
		return fmt.Errorf("API返回错误: %s", apiResp.Message)
	}

	indicators.AHR999 = apiResp.Data.AHR999
	return nil
}

// fetchPrices 获取BTC和ETH价格
func (c *CollectorV2) fetchPrices(indicators *model.MarketIndicators) error {
	// TODO: 接入真实价格API（如CoinGecko、Binance等）
	// 目前使用Mock数据
	indicators.CurrentPriceBTC = 95000.0
	indicators.CurrentPriceETH = 3500.0
	return nil
}

// fetchMVRVZScore 获取MVRV-Z Score
func (c *CollectorV2) fetchMVRVZScore() float64 {
	// TODO: 接入真实MVRV-Z API（如Glassnode等）
	// 目前使用Mock数据：正常范围内的值
	return 2.5
}

// calculateMaMultiplierState 计算MA乘数状态
func (c *CollectorV2) calculateMaMultiplierState(btcPrice float64) model.MaMultiplierState {
	// TODO: 接入真实数据或计算730日均线
	// 简单Mock逻辑：
	// - 价格 < 20000: 熊市底部
	// - 价格 > 150000: 疯牛顶部
	// - 其他: 正常
	if btcPrice > 0 {
		if btcPrice < 20000 {
			return model.MaStateBearBottom
		}
		if btcPrice > 150000 {
			return model.MaStateBullTop
		}
	}
	return model.MaStateNormal
}

// fetchPiCycleStatus 获取Pi周期状态
func (c *CollectorV2) fetchPiCycleStatus() bool {
	// TODO: 接入真实Pi周期数据
	// 目前返回false（未死叉）
	return false
}

// calculateEthRegressionState 计算ETH回归带位置
func (c *CollectorV2) calculateEthRegressionState(ethPrice float64) model.EthRegressionState {
	// TODO: 接入真实回归带数据
	// 简单Mock逻辑：
	// - 价格 < 2000: 低估区
	// - 价格 > 5000: 高估区
	// - 其他: 中间区
	if ethPrice > 0 {
		if ethPrice < 2000 {
			return model.EthRegLower
		}
		if ethPrice > 5000 {
			return model.EthRegUpper
		}
	}
	return model.EthRegMiddle
}
