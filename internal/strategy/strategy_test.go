package strategy

import "testing"

// TestEvaluate_HighLeverage 测试风控规则：高杠杆警告
func TestEvaluate_HighLeverage(t *testing.T) {
	// 杠杆率 > 1.5 时，无论AHR999是多少，都应该返回警告
	testCases := []struct {
		name     string
		ahr999   float64
		leverage float64
	}{
		{"杠杆1.6_抄底区", 0.30, 1.6},
		{"杠杆2.0_定投区", 0.60, 2.0},
		{"杠杆3.0_持有区", 1.0, 3.0},
		{"杠杆1.51_逃顶区", 1.5, 1.51},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decision := Evaluate(tc.ahr999, tc.leverage)
			if decision.Action != ActionWarning {
				t.Errorf("期望 Action=%s, 实际=%s", ActionWarning, decision.Action)
			}
			if decision.AmountFactor != 0 {
				t.Errorf("期望 AmountFactor=0, 实际=%f", decision.AmountFactor)
			}
		})
	}
}

// TestEvaluate_BottomFishing 测试抄底区：AHR999 < 0.45
func TestEvaluate_BottomFishing(t *testing.T) {
	testCases := []struct {
		name   string
		ahr999 float64
	}{
		{"AHR999_0.30", 0.30},
		{"AHR999_0.44", 0.44},
		{"AHR999_0.10", 0.10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decision := Evaluate(tc.ahr999, 1.0) // 正常杠杆
			if decision.Action != ActionGreedyBuy {
				t.Errorf("期望 Action=%s, 实际=%s", ActionGreedyBuy, decision.Action)
			}
			if decision.AmountFactor != 1.5 {
				t.Errorf("期望 AmountFactor=1.5, 实际=%f", decision.AmountFactor)
			}
		})
	}
}

// TestEvaluate_DCA 测试定投区：0.45 <= AHR999 < 0.80
func TestEvaluate_DCA(t *testing.T) {
	testCases := []struct {
		name   string
		ahr999 float64
	}{
		{"AHR999_0.45", 0.45},
		{"AHR999_0.60", 0.60},
		{"AHR999_0.79", 0.79},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decision := Evaluate(tc.ahr999, 1.0) // 正常杠杆
			if decision.Action != ActionDCA {
				t.Errorf("期望 Action=%s, 实际=%s", ActionDCA, decision.Action)
			}
			if decision.AmountFactor != 1.0 {
				t.Errorf("期望 AmountFactor=1.0, 实际=%f", decision.AmountFactor)
			}
		})
	}
}

// TestEvaluate_Hold 测试持有区：0.80 <= AHR999 < 1.2
func TestEvaluate_Hold(t *testing.T) {
	testCases := []struct {
		name   string
		ahr999 float64
	}{
		{"AHR999_0.80", 0.80},
		{"AHR999_1.0", 1.0},
		{"AHR999_1.19", 1.19},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decision := Evaluate(tc.ahr999, 1.0) // 正常杠杆
			if decision.Action != ActionHold {
				t.Errorf("期望 Action=%s, 实际=%s", ActionHold, decision.Action)
			}
			if decision.AmountFactor != 0 {
				t.Errorf("期望 AmountFactor=0, 实际=%f", decision.AmountFactor)
			}
		})
	}
}

// TestEvaluate_Sell 测试逃顶区：AHR999 >= 1.2
func TestEvaluate_Sell(t *testing.T) {
	testCases := []struct {
		name   string
		ahr999 float64
	}{
		{"AHR999_1.2", 1.2},
		{"AHR999_1.5", 1.5},
		{"AHR999_2.0", 2.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decision := Evaluate(tc.ahr999, 1.0) // 正常杠杆
			if decision.Action != ActionSell {
				t.Errorf("期望 Action=%s, 实际=%s", ActionSell, decision.Action)
			}
			if decision.AmountFactor != 0 {
				t.Errorf("期望 AmountFactor=0, 实际=%f", decision.AmountFactor)
			}
		})
	}
}

// TestEvaluate_EdgeCases 测试边界值
func TestEvaluate_EdgeCases(t *testing.T) {
	// 边界值：0.45（定投区下限）
	decision := Evaluate(0.45, 1.0)
	if decision.Action != ActionDCA {
		t.Errorf("AHR999=0.45 应该在定投区, 期望=%s, 实际=%s", ActionDCA, decision.Action)
	}

	// 边界值：0.80（持有区下限）
	decision = Evaluate(0.80, 1.0)
	if decision.Action != ActionHold {
		t.Errorf("AHR999=0.80 应该在持有区, 期望=%s, 实际=%s", ActionHold, decision.Action)
	}

	// 边界值：1.2（逃顶区下限）
	decision = Evaluate(1.2, 1.0)
	if decision.Action != ActionSell {
		t.Errorf("AHR999=1.2 应该在逃顶区, 期望=%s, 实际=%s", ActionSell, decision.Action)
	}

	// 边界值：杠杆率正好1.5（不应触发警告）
	decision = Evaluate(0.60, 1.5)
	if decision.Action != ActionDCA {
		t.Errorf("杠杆=1.5 不应触发警告, 期望=%s, 实际=%s", ActionDCA, decision.Action)
	}
}
