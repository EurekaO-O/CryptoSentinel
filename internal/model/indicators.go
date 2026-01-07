// Package model 定义市场数据模型
package model

import "time"

// MaMultiplierState 两年MA乘数状态枚举
type MaMultiplierState int

const (
	// MaStateNormal 正常状态
	MaStateNormal MaMultiplierState = 0
	// MaStateBearBottom 跌破绿线/熊市底部
	MaStateBearBottom MaMultiplierState = 1
	// MaStateBullTop 突破红线/疯牛顶部
	MaStateBullTop MaMultiplierState = 2
)

// String 返回MA状态的字符串描述
func (s MaMultiplierState) String() string {
	switch s {
	case MaStateNormal:
		return "正常"
	case MaStateBearBottom:
		return "熊市底部"
	case MaStateBullTop:
		return "疯牛顶部"
	default:
		return "未知"
	}
}

// EthRegressionState 以太坊回归带位置
type EthRegressionState string

const (
	// EthRegLower 低估区
	EthRegLower EthRegressionState = "lower"
	// EthRegMiddle 中间区
	EthRegMiddle EthRegressionState = "middle"
	// EthRegUpper 高估区
	EthRegUpper EthRegressionState = "upper"
)

// MarketIndicators 聚合市场指标结构体
type MarketIndicators struct {
	// CurrentPriceBTC 当前BTC价格（USD）
	CurrentPriceBTC float64 `json:"current_price_btc"`

	// CurrentPriceETH 当前ETH价格（USD）
	CurrentPriceETH float64 `json:"current_price_eth"`

	// AHR999 核心囤币指标
	// < 0.45: 抄底区
	// 0.45 - 1.20: 定投区
	// 1.20 - 5.00: 持有区
	// > 5.00: 卖出区
	AHR999 float64 `json:"ahr999"`

	// MVRVZScore 市场冷热指标
	// > 6.0 表示极度过热
	MVRVZScore float64 `json:"mvrv_z_score"`

	// MaMultiplierState 两年MA乘数状态
	// 0: 正常
	// 1: 跌破绿线/熊市底
	// 2: 突破红线/疯牛顶
	MaMultiplierState MaMultiplierState `json:"ma_multiplier_state"`

	// PiCycleCross Pi周期是否死叉（顶部信号）
	PiCycleCross bool `json:"pi_cycle_cross"`

	// EthRegressionState 以太坊回归带位置
	// "lower": 低估区
	// "middle": 中间区
	// "upper": 高估区
	EthRegressionState EthRegressionState `json:"eth_regression_state"`

	// AccountLeverage 当前账户有效杠杆
	AccountLeverage float64 `json:"account_leverage"`

	// Timestamp 数据获取时间
	Timestamp time.Time `json:"timestamp"`

	// Source 数据来源
	Source string `json:"source"`
}

// TradeSignal 交易信号结构体
type TradeSignal struct {
	// ActionBTC BTC操作建议
	ActionBTC string `json:"action_btc"`

	// ActionETH ETH操作建议
	ActionETH string `json:"action_eth"`

	// AmountFactor 资金倍率
	AmountFactor float64 `json:"amount_factor"`

	// WarningMsg 警告消息
	WarningMsg string `json:"warning_msg"`

	// ReportMarkdown Markdown格式的报告
	ReportMarkdown string `json:"report_markdown"`

	// IsHalted 是否触发熔断
	IsHalted bool `json:"is_halted"`
}

// BTC操作常量
const (
	ActionHalt        = "HALT"         // 熔断停止
	ActionSellAlert   = "SELL_ALERT"   // 逃顶警报
	ActionStrongBuy   = "STRONG_BUY"   // 强力买入
	ActionDCABuy      = "DCA_BUY"      // 定投买入
	ActionHold        = "HOLD"         // 持有
	ActionHoldCaution = "HOLD_CAUTION" // 谨慎持有
	ActionSell        = "SELL"         // 卖出
)

// ETH操作常量
const (
	ActionETHBuyHeavy    = "BUY_HEAVY"        // 重仓买入
	ActionETHSellOrSwap  = "SELL_OR_SWAP_BTC" // 卖出或换BTC
	ActionETHFollowBTC   = "FOLLOW_BTC"       // 跟随BTC策略
)
