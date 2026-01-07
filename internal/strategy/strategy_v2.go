// Package strategy 提供升级版多因子投资策略逻辑
package strategy

import (
	"fmt"
	"time"

	"CryptoSentinel/internal/model"
)

// 阈值常量
const (
	// 杠杆率阈值
	MaxLeverage = 1.5

	// AHR999 阈值
	AHR999StrongBuy = 0.45
	AHR999DCAUpper  = 1.20
	AHR999HoldUpper = 5.00

	// MVRV-Z 阈值
	MVRVOverheat = 6.0
)

// EvaluateV2 升级版策略评估函数
// 严格按照"风控 -> 信号 -> 执行"优先级执行
func EvaluateV2(indicators *model.MarketIndicators) *model.TradeSignal {
	signal := &model.TradeSignal{
		ActionBTC:    model.ActionHold,
		ActionETH:    model.ActionETHFollowBTC,
		AmountFactor: 0,
		IsHalted:     false,
	}

	// ========== 第一步：风控熔断 (Safety First) ==========

	// 1.1 杠杆检查
	if indicators.AccountLeverage > MaxLeverage {
		signal.ActionBTC = model.ActionHalt
		signal.ActionETH = model.ActionHalt
		signal.WarningMsg = fmt.Sprintf("⚠️ 杠杆过高(%.2fx > %.1fx)！停止买入，请补充保证金！",
			indicators.AccountLeverage, MaxLeverage)
		signal.IsHalted = true
		signal.ReportMarkdown = generateReport(indicators, signal)
		return signal
	}

	// 1.2 逃顶检查
	if indicators.PiCycleCross || indicators.MaMultiplierState == model.MaStateBullTop {
		signal.ActionBTC = model.ActionSellAlert
		signal.ActionETH = model.ActionSellAlert
		signal.WarningMsg = "🚨 逃顶信号触发！Pi周期死叉或突破两年红线！禁止买入！"
		signal.IsHalted = true
		signal.ReportMarkdown = generateReport(indicators, signal)
		return signal
	}

	// ========== 第二步：计算 BTC 基础策略 (Based on AHR999) ==========

	btcAllowBuy := false

	if indicators.AHR999 < AHR999StrongBuy {
		// 抄底区：强力买入
		signal.ActionBTC = model.ActionStrongBuy
		signal.AmountFactor = 1.5
		btcAllowBuy = true
	} else if indicators.AHR999 >= AHR999StrongBuy && indicators.AHR999 < AHR999DCAUpper {
		// 定投区：定投买入
		signal.ActionBTC = model.ActionDCABuy
		signal.AmountFactor = 1.0
		btcAllowBuy = true
	} else if indicators.AHR999 >= AHR999DCAUpper && indicators.AHR999 < AHR999HoldUpper {
		// 持有区：只囤U
		signal.ActionBTC = model.ActionHold
		signal.AmountFactor = 0
	} else {
		// 卖出区
		signal.ActionBTC = model.ActionSell
		signal.AmountFactor = 0
	}

	// ========== 第三步：MVRV 辅助验证 ==========

	if btcAllowBuy && indicators.MVRVZScore > MVRVOverheat {
		// MVRV-Z 极度过热，强制改为谨慎持有
		signal.ActionBTC = model.ActionHoldCaution
		signal.AmountFactor = 0
		signal.WarningMsg = fmt.Sprintf("⚠️ MVRV-Z Score (%.2f) 极度过热，暂停买入！", indicators.MVRVZScore)
		btcAllowBuy = false
	}

	// ========== 第四步：ETH 独立策略 (Eth Regression) ==========

	switch indicators.EthRegressionState {
	case model.EthRegLower:
		// 低估区且BTC允许买入
		if btcAllowBuy {
			signal.ActionETH = model.ActionETHBuyHeavy
		} else {
			signal.ActionETH = model.ActionETHFollowBTC
		}
	case model.EthRegUpper:
		// 高估区：卖出或换BTC
		signal.ActionETH = model.ActionETHSellOrSwap
	default:
		// 中间区：跟随BTC
		signal.ActionETH = model.ActionETHFollowBTC
	}

	// ========== 第五步：生成报告 ==========

	signal.ReportMarkdown = generateReport(indicators, signal)

	return signal
}

// generateReport 生成Markdown格式的报告
func generateReport(indicators *model.MarketIndicators, signal *model.TradeSignal) string {
	// 杠杆状态
	leverageStatus := "✅ 安全"
	if indicators.AccountLeverage > MaxLeverage {
		leverageStatus = "❌ 危险"
	} else if indicators.AccountLeverage > 1.2 {
		leverageStatus = "⚠️ 警戒"
	}

	// 逃顶信号
	piStatus := "✅ 正常"
	if indicators.PiCycleCross {
		piStatus = "❌ 死叉"
	}

	maStatus := indicators.MaMultiplierState.String()

	// AHR999区间描述
	ahr999Desc := getAHR999Description(indicators.AHR999)

	// ETH位置描述
	ethPosDesc := getEthPositionDescription(indicators.EthRegressionState)

	// BTC操作描述
	btcActionDesc := getBTCActionDescription(signal.ActionBTC)

	// ETH操作描述
	ethActionDesc := getETHActionDescription(signal.ActionETH)

	// 警告信息
	warningSection := ""
	if signal.WarningMsg != "" {
		warningSection = fmt.Sprintf("\n⚠️ **警告**: %s\n", signal.WarningMsg)
	}

	return fmt.Sprintf(`🛡️ **CryptoSentinel 周报** [%s]
%s
**1. 风控检查**
- 杠杆率: %.2fx (%s)
- 逃顶信号: Pi周期[%s] / MA状态[%s]

**2. 核心指标**
- 📏 AHR999: %.4f -> %s
- 🌡️ MVRV-Z: %.2f
- 💎 ETH位置: %s

**3. 执行建议**
- 🚀 **BTC 操作**: %s
- 💰 **倍率**: %.1fx
- Ξ **ETH 操作**: %s

_"看着资产像树苗一样慢慢长高，本身就是一件枯燥的事情。"_`,
		time.Now().Format("2006-01-02"),
		warningSection,
		indicators.AccountLeverage, leverageStatus,
		piStatus, maStatus,
		indicators.AHR999, ahr999Desc,
		indicators.MVRVZScore,
		ethPosDesc,
		btcActionDesc,
		signal.AmountFactor,
		ethActionDesc,
	)
}

// getAHR999Description 获取AHR999区间描述
func getAHR999Description(ahr999 float64) string {
	if ahr999 < AHR999StrongBuy {
		return "🟢 抄底区"
	} else if ahr999 < AHR999DCAUpper {
		return "🔵 定投区"
	} else if ahr999 < AHR999HoldUpper {
		return "🟡 持有区"
	}
	return "🔴 逃顶区"
}

// getEthPositionDescription 获取ETH位置描述
func getEthPositionDescription(state model.EthRegressionState) string {
	switch state {
	case model.EthRegLower:
		return "🟢 低估区"
	case model.EthRegMiddle:
		return "🟡 中间区"
	case model.EthRegUpper:
		return "🔴 高估区"
	default:
		return "❓ 未知"
	}
}

// getBTCActionDescription 获取BTC操作描述
func getBTCActionDescription(action string) string {
	switch action {
	case model.ActionHalt:
		return "🛑 熔断停止"
	case model.ActionSellAlert:
		return "🚨 逃顶警报"
	case model.ActionStrongBuy:
		return "💪 贪婪买入"
	case model.ActionDCABuy:
		return "📈 正常定投"
	case model.ActionHold:
		return "✋ 持有观望"
	case model.ActionHoldCaution:
		return "⚠️ 谨慎持有"
	case model.ActionSell:
		return "📉 逐步卖出"
	default:
		return action
	}
}

// getETHActionDescription 获取ETH操作描述
func getETHActionDescription(action string) string {
	switch action {
	case model.ActionHalt:
		return "🛑 熔断停止"
	case model.ActionSellAlert:
		return "🚨 逃顶警报"
	case model.ActionETHBuyHeavy:
		return "💪 重仓买入"
	case model.ActionETHSellOrSwap:
		return "🔄 卖出/换BTC"
	case model.ActionETHFollowBTC:
		return "👉 跟随BTC"
	default:
		return action
	}
}
