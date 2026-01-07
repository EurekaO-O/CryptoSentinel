// Package main 演示模式 - 展示策略引擎效果
package main

import (
	"fmt"
	"time"

	"CryptoSentinel/internal/model"
	"CryptoSentinel/internal/strategy"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("  CryptoSentinel 策略演示")
	fmt.Println("========================================")
	fmt.Println()

	// 演示场景1：正常定投
	demo("场景1: 正常定投", &model.MarketIndicators{
		CurrentPriceBTC:    95000,
		CurrentPriceETH:    3500,
		AHR999:             0.70,
		MVRVZScore:         2.5,
		MaMultiplierState:  model.MaStateNormal,
		PiCycleCross:       false,
		EthRegressionState: model.EthRegMiddle,
		AccountLeverage:    1.0,
		Timestamp:          time.Now(),
	})

	// 演示场景2：抄底机会
	demo("场景2: 抄底机会 (AHR999 < 0.45)", &model.MarketIndicators{
		CurrentPriceBTC:    45000,
		CurrentPriceETH:    1800,
		AHR999:             0.30,
		MVRVZScore:         0.5,
		MaMultiplierState:  model.MaStateNormal,
		PiCycleCross:       false,
		EthRegressionState: model.EthRegLower,
		AccountLeverage:    1.0,
		Timestamp:          time.Now(),
	})

	// 演示场景3：高杠杆熔断
	demo("场景3: 高杠杆熔断", &model.MarketIndicators{
		CurrentPriceBTC:    95000,
		CurrentPriceETH:    3500,
		AHR999:             0.70,
		MVRVZScore:         2.5,
		MaMultiplierState:  model.MaStateNormal,
		PiCycleCross:       false,
		EthRegressionState: model.EthRegMiddle,
		AccountLeverage:    2.0, // 杠杆过高
		Timestamp:          time.Now(),
	})

	// 演示场景4：逃顶信号
	demo("场景4: 逃顶信号 (Pi周期死叉)", &model.MarketIndicators{
		CurrentPriceBTC:    150000,
		CurrentPriceETH:    8000,
		AHR999:             3.5,
		MVRVZScore:         5.5,
		MaMultiplierState:  model.MaStateNormal,
		PiCycleCross:       true, // Pi周期死叉
		EthRegressionState: model.EthRegUpper,
		AccountLeverage:    1.0,
		Timestamp:          time.Now(),
	})

	// 演示场景5：MVRV过热
	demo("场景5: MVRV过热 (谨慎持有)", &model.MarketIndicators{
		CurrentPriceBTC:    120000,
		CurrentPriceETH:    6000,
		AHR999:             0.70,
		MVRVZScore:         7.0, // 极度过热
		MaMultiplierState:  model.MaStateNormal,
		PiCycleCross:       false,
		EthRegressionState: model.EthRegMiddle,
		AccountLeverage:    1.0,
		Timestamp:          time.Now(),
	})
}

func demo(title string, indicators *model.MarketIndicators) {
	fmt.Printf("--- %s ---\n", title)
	fmt.Printf("输入: AHR999=%.2f, 杠杆=%.1fx, MVRV-Z=%.1f, Pi死叉=%v\n",
		indicators.AHR999, indicators.AccountLeverage, indicators.MVRVZScore, indicators.PiCycleCross)
	fmt.Println()

	signal := strategy.EvaluateV2(indicators)

	fmt.Println(signal.ReportMarkdown)
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println()
}
