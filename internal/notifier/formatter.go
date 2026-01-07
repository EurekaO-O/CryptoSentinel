// Package notifier 提供消息通知和格式化功能
package notifier

import (
	"bytes"
	"text/template"
	"time"

	"CryptoSentinel/internal/model"
)

// ReportData 报告模板数据结构
type ReportData struct {
	Date           string
	LeverageValue  float64
	LeverageStatus string
	PiStatus       string
	MaStatus       string
	AHR999Value    float64
	AHR999Desc     string
	MVRVZValue     float64
	ETHPosition    string
	BTCAction      string
	AmountFactor   float64
	ETHAction      string
	WarningMsg     string
	HasWarning     bool
}

// 报告模板
const reportTemplate = `🛡️ **CryptoSentinel 周报** [{{.Date}}]
{{if .HasWarning}}
⚠️ **警告**: {{.WarningMsg}}
{{end}}
**1. 风控检查**
- 杠杆率: {{printf "%.2f" .LeverageValue}}x ({{.LeverageStatus}})
- 逃顶信号: Pi周期[{{.PiStatus}}] / MA状态[{{.MaStatus}}]

**2. 核心指标**
- 📏 AHR999: {{printf "%.4f" .AHR999Value}} -> {{.AHR999Desc}}
- 🌡️ MVRV-Z: {{printf "%.2f" .MVRVZValue}}
- 💎 ETH位置: {{.ETHPosition}}

**3. 执行建议**
- 🚀 **BTC 操作**: {{.BTCAction}}
- 💰 **倍率**: {{printf "%.1f" .AmountFactor}}x
- Ξ **ETH 操作**: {{.ETHAction}}

_"看着资产像树苗一样慢慢长高，本身就是一件枯燥的事情。"_`

// FormatReport 使用模板格式化报告
func FormatReport(indicators *model.MarketIndicators, signal *model.TradeSignal) (string, error) {
	tmpl, err := template.New("report").Parse(reportTemplate)
	if err != nil {
		return "", err
	}

	data := ReportData{
		Date:           time.Now().Format("2006-01-02"),
		LeverageValue:  indicators.AccountLeverage,
		LeverageStatus: getLeverageStatus(indicators.AccountLeverage),
		PiStatus:       getPiStatus(indicators.PiCycleCross),
		MaStatus:       indicators.MaMultiplierState.String(),
		AHR999Value:    indicators.AHR999,
		AHR999Desc:     getAHR999Desc(indicators.AHR999),
		MVRVZValue:     indicators.MVRVZScore,
		ETHPosition:    getETHPosition(indicators.EthRegressionState),
		BTCAction:      getBTCAction(signal.ActionBTC),
		AmountFactor:   signal.AmountFactor,
		ETHAction:      getETHAction(signal.ActionETH),
		WarningMsg:     signal.WarningMsg,
		HasWarning:     signal.WarningMsg != "",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// getLeverageStatus 获取杠杆状态描述
func getLeverageStatus(leverage float64) string {
	if leverage > 1.5 {
		return "❌ 危险"
	} else if leverage > 1.2 {
		return "⚠️ 警戒"
	}
	return "✅ 安全"
}

// getPiStatus 获取Pi周期状态
func getPiStatus(cross bool) string {
	if cross {
		return "❌ 死叉"
	}
	return "✅ 正常"
}

// getAHR999Desc 获取AHR999区间描述
func getAHR999Desc(ahr999 float64) string {
	if ahr999 < 0.45 {
		return "🟢 抄底区"
	} else if ahr999 < 1.20 {
		return "🔵 定投区"
	} else if ahr999 < 5.00 {
		return "🟡 持有区"
	}
	return "🔴 逃顶区"
}

// getETHPosition 获取ETH位置描述
func getETHPosition(state model.EthRegressionState) string {
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

// getBTCAction 获取BTC操作描述
func getBTCAction(action string) string {
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

// getETHAction 获取ETH操作描述
func getETHAction(action string) string {
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
