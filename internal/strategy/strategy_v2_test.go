package strategy

import (
	"strings"
	"testing"
	"time"

	"CryptoSentinel/internal/model"
)

// 创建基础测试指标
func newTestIndicators() *model.MarketIndicators {
	return &model.MarketIndicators{
		CurrentPriceBTC:    95000,
		CurrentPriceETH:    3500,
		AHR999:             0.70,
		MVRVZScore:         2.5,
		MaMultiplierState:  model.MaStateNormal,
		PiCycleCross:       false,
		EthRegressionState: model.EthRegMiddle,
		AccountLeverage:    1.0,
		Timestamp:          time.Now(),
	}
}

// TestEvaluateV2_NormalDCA 测试场景1：正常定投场景
func TestEvaluateV2_NormalDCA(t *testing.T) {
	indicators := newTestIndicators()
	// AHR999 = 0.70 在定投区 (0.45 - 1.20)
	indicators.AHR999 = 0.70
	indicators.AccountLeverage = 1.0

	signal := EvaluateV2(indicators)

	// 验证BTC操作
	if signal.ActionBTC != model.ActionDCABuy {
		t.Errorf("正常定投场景: 期望BTC Action=%s, 实际=%s", model.ActionDCABuy, signal.ActionBTC)
	}

	// 验证资金倍率
	if signal.AmountFactor != 1.0 {
		t.Errorf("正常定投场景: 期望AmountFactor=1.0, 实际=%f", signal.AmountFactor)
	}

	// 验证未触发熔断
	if signal.IsHalted {
		t.Error("正常定投场景: 不应该触发熔断")
	}

	// 验证ETH跟随BTC
	if signal.ActionETH != model.ActionETHFollowBTC {
		t.Errorf("正常定投场景: 期望ETH Action=%s, 实际=%s", model.ActionETHFollowBTC, signal.ActionETH)
	}

	// 验证报告不为空
	if signal.ReportMarkdown == "" {
		t.Error("正常定投场景: 报告不应为空")
	}
}

// TestEvaluateV2_HighLeverageHalt 测试场景2：高杠杆熔断场景
func TestEvaluateV2_HighLeverageHalt(t *testing.T) {
	indicators := newTestIndicators()
	// 设置高杠杆
	indicators.AccountLeverage = 2.0

	signal := EvaluateV2(indicators)

	// 验证触发熔断
	if !signal.IsHalted {
		t.Error("高杠杆场景: 应该触发熔断")
	}

	// 验证BTC操作为HALT
	if signal.ActionBTC != model.ActionHalt {
		t.Errorf("高杠杆场景: 期望BTC Action=%s, 实际=%s", model.ActionHalt, signal.ActionBTC)
	}

	// 验证ETH操作为HALT
	if signal.ActionETH != model.ActionHalt {
		t.Errorf("高杠杆场景: 期望ETH Action=%s, 实际=%s", model.ActionHalt, signal.ActionETH)
	}

	// 验证资金倍率为0
	if signal.AmountFactor != 0 {
		t.Errorf("高杠杆场景: 期望AmountFactor=0, 实际=%f", signal.AmountFactor)
	}

	// 验证警告消息包含杠杆信息
	if !strings.Contains(signal.WarningMsg, "杠杆过高") {
		t.Errorf("高杠杆场景: 警告消息应包含'杠杆过高', 实际=%s", signal.WarningMsg)
	}
}

// TestEvaluateV2_BottomFishing 测试场景3：抄底场景
func TestEvaluateV2_BottomFishing(t *testing.T) {
	indicators := newTestIndicators()
	// AHR999 < 0.45 抄底区
	indicators.AHR999 = 0.30
	indicators.AccountLeverage = 1.0
	// ETH也在低估区
	indicators.EthRegressionState = model.EthRegLower

	signal := EvaluateV2(indicators)

	// 验证BTC操作为强力买入
	if signal.ActionBTC != model.ActionStrongBuy {
		t.Errorf("抄底场景: 期望BTC Action=%s, 实际=%s", model.ActionStrongBuy, signal.ActionBTC)
	}

	// 验证资金倍率为1.5x
	if signal.AmountFactor != 1.5 {
		t.Errorf("抄底场景: 期望AmountFactor=1.5, 实际=%f", signal.AmountFactor)
	}

	// 验证未触发熔断
	if signal.IsHalted {
		t.Error("抄底场景: 不应该触发熔断")
	}

	// 验证ETH重仓买入（因为在低估区且BTC允许买入）
	if signal.ActionETH != model.ActionETHBuyHeavy {
		t.Errorf("抄底场景: 期望ETH Action=%s, 实际=%s", model.ActionETHBuyHeavy, signal.ActionETH)
	}
}

// TestEvaluateV2_PiCycleCross 测试Pi周期死叉逃顶
func TestEvaluateV2_PiCycleCross(t *testing.T) {
	indicators := newTestIndicators()
	// Pi周期死叉
	indicators.PiCycleCross = true

	signal := EvaluateV2(indicators)

	// 验证触发熔断
	if !signal.IsHalted {
		t.Error("Pi周期死叉场景: 应该触发熔断")
	}

	// 验证操作为卖出警报
	if signal.ActionBTC != model.ActionSellAlert {
		t.Errorf("Pi周期死叉场景: 期望BTC Action=%s, 实际=%s", model.ActionSellAlert, signal.ActionBTC)
	}

	// 验证警告消息
	if !strings.Contains(signal.WarningMsg, "逃顶信号") {
		t.Errorf("Pi周期死叉场景: 警告消息应包含'逃顶信号', 实际=%s", signal.WarningMsg)
	}
}

// TestEvaluateV2_MaBullTop 测试MA突破红线逃顶
func TestEvaluateV2_MaBullTop(t *testing.T) {
	indicators := newTestIndicators()
	// MA突破红线
	indicators.MaMultiplierState = model.MaStateBullTop

	signal := EvaluateV2(indicators)

	// 验证触发熔断
	if !signal.IsHalted {
		t.Error("MA疯牛顶场景: 应该触发熔断")
	}

	// 验证操作为卖出警报
	if signal.ActionBTC != model.ActionSellAlert {
		t.Errorf("MA疯牛顶场景: 期望BTC Action=%s, 实际=%s", model.ActionSellAlert, signal.ActionBTC)
	}
}

// TestEvaluateV2_MVRVOverheat 测试MVRV过热谨慎持有
func TestEvaluateV2_MVRVOverheat(t *testing.T) {
	indicators := newTestIndicators()
	// AHR999在定投区，但MVRV极度过热
	indicators.AHR999 = 0.70
	indicators.MVRVZScore = 7.0

	signal := EvaluateV2(indicators)

	// 验证操作改为谨慎持有
	if signal.ActionBTC != model.ActionHoldCaution {
		t.Errorf("MVRV过热场景: 期望BTC Action=%s, 实际=%s", model.ActionHoldCaution, signal.ActionBTC)
	}

	// 验证资金倍率为0
	if signal.AmountFactor != 0 {
		t.Errorf("MVRV过热场景: 期望AmountFactor=0, 实际=%f", signal.AmountFactor)
	}

	// 验证警告消息
	if !strings.Contains(signal.WarningMsg, "MVRV-Z Score") {
		t.Errorf("MVRV过热场景: 警告消息应包含'MVRV-Z Score', 实际=%s", signal.WarningMsg)
	}
}

// TestEvaluateV2_ETHUpperRegression 测试ETH高估区卖出
func TestEvaluateV2_ETHUpperRegression(t *testing.T) {
	indicators := newTestIndicators()
	indicators.AHR999 = 0.70
	// ETH在高估区
	indicators.EthRegressionState = model.EthRegUpper

	signal := EvaluateV2(indicators)

	// 验证ETH操作为卖出或换BTC
	if signal.ActionETH != model.ActionETHSellOrSwap {
		t.Errorf("ETH高估区场景: 期望ETH Action=%s, 实际=%s", model.ActionETHSellOrSwap, signal.ActionETH)
	}
}

// TestEvaluateV2_HoldZone 测试持有区
func TestEvaluateV2_HoldZone(t *testing.T) {
	indicators := newTestIndicators()
	// AHR999在持有区 (1.20 - 5.00)
	indicators.AHR999 = 2.5

	signal := EvaluateV2(indicators)

	// 验证BTC操作为持有
	if signal.ActionBTC != model.ActionHold {
		t.Errorf("持有区场景: 期望BTC Action=%s, 实际=%s", model.ActionHold, signal.ActionBTC)
	}

	// 验证资金倍率为0
	if signal.AmountFactor != 0 {
		t.Errorf("持有区场景: 期望AmountFactor=0, 实际=%f", signal.AmountFactor)
	}
}

// TestEvaluateV2_SellZone 测试卖出区
func TestEvaluateV2_SellZone(t *testing.T) {
	indicators := newTestIndicators()
	// AHR999 > 5.00 卖出区
	indicators.AHR999 = 6.0

	signal := EvaluateV2(indicators)

	// 验证BTC操作为卖出
	if signal.ActionBTC != model.ActionSell {
		t.Errorf("卖出区场景: 期望BTC Action=%s, 实际=%s", model.ActionSell, signal.ActionBTC)
	}

	// 验证资金倍率为0
	if signal.AmountFactor != 0 {
		t.Errorf("卖出区场景: 期望AmountFactor=0, 实际=%f", signal.AmountFactor)
	}
}

// TestGenerateReport 测试报告生成
func TestGenerateReport(t *testing.T) {
	indicators := newTestIndicators()
	signal := EvaluateV2(indicators)

	// 验证报告包含必要元素
	report := signal.ReportMarkdown

	requiredElements := []string{
		"CryptoSentinel 周报",
		"风控检查",
		"核心指标",
		"执行建议",
		"AHR999",
		"MVRV-Z",
		"ETH位置",
		"BTC 操作",
		"ETH 操作",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(report, elem) {
			t.Errorf("报告应包含 '%s'", elem)
		}
	}
}
