// Package calculator 提供加密货币指标计算功能
package calculator

import (
	"fmt"
	"time"
)

// MA730State 两年MA乘数状态
type MA730State int

const (
	// MA730Normal 正常区间
	MA730Normal MA730State = 0
	// MA730Bottom 抄底区（价格低于MA730）
	MA730Bottom MA730State = 1
	// MA730Top 逃顶区（价格高于MA730*5）
	MA730Top MA730State = 2
)

// String 返回状态描述
func (s MA730State) String() string {
	switch s {
	case MA730Normal:
		return "正常"
	case MA730Bottom:
		return "抄底区(绿线下方)"
	case MA730Top:
		return "逃顶区(红线上方)"
	default:
		return "未知"
	}
}

// MA730Result MA730计算结果
type MA730Result struct {
	// CurrentPrice 当前BTC价格
	CurrentPrice float64
	// MA730 730日移动平均线
	MA730 float64
	// MA730x5 MA730 * 5 (红线/逃顶线)
	MA730x5 float64
	// Multiplier 当前价格相对于MA730的倍数
	Multiplier float64
	// State 当前状态
	State MA730State
	// Timestamp 计算时间
	Timestamp time.Time
}

// MA730Calculator 两年MA乘数计算器
type MA730Calculator struct {
	*AHR999Calculator // 复用K线获取逻辑
}

// NewMA730Calculator 创建MA730计算器
func NewMA730Calculator() *MA730Calculator {
	return &MA730Calculator{
		AHR999Calculator: NewAHR999Calculator(),
	}
}

// NewMA730CalculatorWithProxy 创建带代理的MA730计算器
func NewMA730CalculatorWithProxy(proxyAddr string) *MA730Calculator {
	return &MA730Calculator{
		AHR999Calculator: NewAHR999CalculatorWithProxy(proxyAddr),
	}
}

// Calculate 计算两年MA乘数指标
func (c *MA730Calculator) Calculate() (*MA730Result, error) {
	// 获取过去730天的K线数据
	// 注意：Binance API limit最大为1000，730天在范围内
	prices, err := c.fetchKlines(730)
	if err != nil {
		return nil, fmt.Errorf("获取K线数据失败: %w", err)
	}

	if len(prices) < 730 {
		return nil, fmt.Errorf("K线数据不足，需要730条，实际获取%d条", len(prices))
	}

	// 当前价格
	currentPrice := prices[len(prices)-1]

	// 计算730日简单移动平均线
	ma730 := calculateSMA(prices)

	// MA730 * 5 (逃顶线)
	ma730x5 := ma730 * 5

	// 计算倍数
	multiplier := currentPrice / ma730

	// 判断状态
	var state MA730State
	if currentPrice < ma730 {
		state = MA730Bottom // 价格低于绿线，抄底区
	} else if currentPrice > ma730x5 {
		state = MA730Top // 价格高于红线，逃顶区
	} else {
		state = MA730Normal // 正常区间
	}

	return &MA730Result{
		CurrentPrice: currentPrice,
		MA730:        ma730,
		MA730x5:      ma730x5,
		Multiplier:   multiplier,
		State:        state,
		Timestamp:    time.Now(),
	}, nil
}

// calculateSMA 计算简单移动平均线
func calculateSMA(prices []float64) float64 {
	if len(prices) == 0 {
		return 0
	}

	var sum float64
	for _, price := range prices {
		sum += price
	}

	return sum / float64(len(prices))
}
