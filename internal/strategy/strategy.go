// Package strategy 提供投资策略逻辑
package strategy

import "fmt"

// Action 投资动作类型
type Action string

const (
	// ActionWarning 风控警告
	ActionWarning Action = "警告"
	// ActionGreedyBuy 贪婪买入（抄底）
	ActionGreedyBuy Action = "贪婪买入"
	// ActionDCA 定投
	ActionDCA Action = "定投"
	// ActionHold 持有观望
	ActionHold Action = "持有观望"
	// ActionSell 卖出
	ActionSell Action = "卖出"
)

// Decision 策略决策结果
type Decision struct {
	// Action 建议动作
	Action Action
	// AmountFactor 资金倍率（相对于基准定投金额）
	AmountFactor float64
	// Message 决策理由说明
	Message string
}

// AHR999Thresholds AHR999阈值常量
const (
	// ThresholdBottomFishing 抄底区阈值
	ThresholdBottomFishing = 0.45
	// ThresholdDCAUpper 定投区上限
	ThresholdDCAUpper = 0.80
	// ThresholdHoldUpper 持有区上限
	ThresholdHoldUpper = 1.2
	// ThresholdMaxLeverage 最大杠杆率
	ThresholdMaxLeverage = 1.5
)

// Evaluate 根据AHR999指数和杠杆率评估投资策略
// ahr999: 当前AHR999指数
// leverage: 当前账户杠杆率
func Evaluate(ahr999 float64, leverage float64) *Decision {
	// 规则1: 风控规则（最高优先级）
	// 如果杠杆率 > 1.5，发出警告
	if leverage > ThresholdMaxLeverage {
		return &Decision{
			Action:       ActionWarning,
			AmountFactor: 0,
			Message:      fmt.Sprintf("杠杆过高 (%.2f > %.2f)，停止买入！请先降低杠杆率。", leverage, ThresholdMaxLeverage),
		}
	}

	// 规则2: 抄底区
	// 如果 AHR999 < 0.45，贪婪买入
	if ahr999 < ThresholdBottomFishing {
		return &Decision{
			Action:       ActionGreedyBuy,
			AmountFactor: 1.5,
			Message:      fmt.Sprintf("AHR999 = %.4f，处于抄底区 (< %.2f)，建议贪婪买入，资金倍率 1.5x", ahr999, ThresholdBottomFishing),
		}
	}

	// 规则3: 定投区
	// 如果 AHR999 在 0.45 到 0.80 之间，正常定投
	if ahr999 >= ThresholdBottomFishing && ahr999 < ThresholdDCAUpper {
		return &Decision{
			Action:       ActionDCA,
			AmountFactor: 1.0,
			Message:      fmt.Sprintf("AHR999 = %.4f，处于定投区 (%.2f - %.2f)，建议正常定投，资金倍率 1.0x", ahr999, ThresholdBottomFishing, ThresholdDCAUpper),
		}
	}

	// 规则4: 持有区
	// 如果 AHR999 在 0.80 到 1.2 之间，持有观望
	if ahr999 >= ThresholdDCAUpper && ahr999 < ThresholdHoldUpper {
		return &Decision{
			Action:       ActionHold,
			AmountFactor: 0,
			Message:      fmt.Sprintf("AHR999 = %.4f，处于持有区 (%.2f - %.2f)，建议持有观望，暂停定投", ahr999, ThresholdDCAUpper, ThresholdHoldUpper),
		}
	}

	// 规则5: 逃顶区
	// 如果 AHR999 > 1.2，卖出
	return &Decision{
		Action:       ActionSell,
		AmountFactor: 0,
		Message:      fmt.Sprintf("AHR999 = %.4f，处于逃顶区 (> %.2f)，建议卖出", ahr999, ThresholdHoldUpper),
	}
}
