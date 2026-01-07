// Package main 多指标集成测试
package main

import (
	"fmt"
	"log"

	"CryptoSentinel/internal/calculator"
	"CryptoSentinel/internal/fetcher"
	"CryptoSentinel/internal/model"
	"CryptoSentinel/internal/notifier"
	"CryptoSentinel/internal/strategy"
)

func main() {
	proxyAddr := "127.0.0.1:10809"

	fmt.Println("🚀 CryptoSentinel 多指标集成测试")
	fmt.Println("================================")
	fmt.Println()

	// 1. 计算 AHR999
	fmt.Println("📊 [1/3] 计算 AHR999...")
	ahr999Calc := calculator.NewAHR999CalculatorWithProxy(proxyAddr)
	ahr999Result, err := ahr999Calc.Calculate()
	if err != nil {
		log.Printf("⚠️ AHR999计算失败: %v", err)
	} else {
		fmt.Printf("   ✅ AHR999: %.4f (当前价格: $%.0f)\n", ahr999Result.AHR999, ahr999Result.CurrentPrice)
	}

	// 2. 计算 MA730 (两年MA乘数)
	fmt.Println("📊 [2/3] 计算 MA730 (两年MA乘数)...")
	ma730Calc := calculator.NewMA730CalculatorWithProxy(proxyAddr)
	ma730Result, err := ma730Calc.Calculate()
	if err != nil {
		log.Printf("⚠️ MA730计算失败: %v", err)
	} else {
		fmt.Printf("   ✅ MA730: $%.0f | 当前倍数: %.2fx | 状态: %s\n",
			ma730Result.MA730, ma730Result.Multiplier, ma730Result.State.String())
	}

	// 3. 获取 MVRV
	fmt.Println("📊 [3/3] 获取 MVRV...")
	mvrvFetcher := fetcher.NewMVRVFetcherWithProxy(proxyAddr)
	mvrvResult, err := mvrvFetcher.Fetch()
	if err != nil {
		log.Printf("⚠️ MVRV获取失败: %v", err)
	} else {
		fmt.Printf("   ✅ MVRV: %.4f | Z-Score: %.2f\n", mvrvResult.MVRV, mvrvResult.ZScore)
	}

	fmt.Println()
	fmt.Println("================================")
	fmt.Println("📋 汇总报告")
	fmt.Println("================================")

	// 汇总显示
	if ahr999Result != nil {
		fmt.Printf("📈 BTC 价格:    $%.0f\n", ahr999Result.CurrentPrice)
		fmt.Printf("🎯 AHR999:      %.4f (%s)\n", ahr999Result.AHR999, getAHR999Zone(ahr999Result.AHR999))
	}
	if ma730Result != nil {
		fmt.Printf("📐 MA730:       $%.0f (倍数: %.2fx)\n", ma730Result.MA730, ma730Result.Multiplier)
		fmt.Printf("🚦 MA状态:      %s\n", ma730Result.State.String())
	}
	if mvrvResult != nil {
		fmt.Printf("🌡️ MVRV:        %.4f\n", mvrvResult.MVRV)
		fmt.Printf("📊 Z-Score:     %.2f\n", mvrvResult.ZScore)
	}

	fmt.Println()

	// 构建完整指标并发送Telegram
	fmt.Println("📤 正在发送完整报告到 Telegram...")

	// 确定MA乘数状态
	maState := model.MaStateNormal
	if ma730Result != nil {
		switch ma730Result.State {
		case calculator.MA730Bottom:
			maState = model.MaStateBearBottom
		case calculator.MA730Top:
			maState = model.MaStateBullTop
		}
	}

	// 构建指标
	currentPrice := 0.0
	ahr999Value := 0.0
	if ahr999Result != nil {
		currentPrice = ahr999Result.CurrentPrice
		ahr999Value = ahr999Result.AHR999
	}

	mvrvZScore := 2.5 // 默认值
	if mvrvResult != nil {
		mvrvZScore = mvrvResult.ZScore
	}

	indicators := &model.MarketIndicators{
		CurrentPriceBTC:    currentPrice,
		CurrentPriceETH:    0,
		AHR999:             ahr999Value,
		MVRVZScore:         mvrvZScore,
		MaMultiplierState:  maState,
		PiCycleCross:       false, // 暂未实现
		EthRegressionState: model.EthRegMiddle,
		AccountLeverage:    1.0,
		Source:             "Binance + CoinMetrics (自主计算)",
	}

	// 执行策略
	signal := strategy.EvaluateV2(indicators)

	// 生成详细报告
	detailReport := signal.ReportMarkdown

	// 添加计算详情
	if ahr999Result != nil {
		detailReport += fmt.Sprintf("\n\n📊 *AHR999 计算详情*\n")
		detailReport += fmt.Sprintf("- 200日定投成本: $%.0f\n", ahr999Result.DCA200Cost)
		detailReport += fmt.Sprintf("- 指数增长估值: $%.0f\n", ahr999Result.ExponentialValue)
	}

	if ma730Result != nil {
		detailReport += fmt.Sprintf("\n📐 *MA730 (两年MA乘数)*\n")
		detailReport += fmt.Sprintf("- 730日均线: $%.0f\n", ma730Result.MA730)
		detailReport += fmt.Sprintf("- 逃顶线(5x): $%.0f\n", ma730Result.MA730x5)
		detailReport += fmt.Sprintf("- 当前倍数: %.2fx\n", ma730Result.Multiplier)
	}

	if mvrvResult != nil {
		detailReport += fmt.Sprintf("\n🌡️ *MVRV*\n")
		detailReport += fmt.Sprintf("- MVRV: %.4f\n", mvrvResult.MVRV)
		detailReport += fmt.Sprintf("- 来源: %s\n", mvrvResult.Source)
	}

	detailReport += "\n_数据全部自主计算/获取，不依赖第三方付费API_"

	// 发送
	botToken := "8530550957:AAE0popkGmCSD6hEsZaGuybOZR5w4CCEmew"
	chatID := "7150834288"
	tg := notifier.NewTelegramNotifierWithProxy(botToken, chatID, proxyAddr)

	if err := tg.Send(detailReport); err != nil {
		log.Fatalf("发送失败: %v", err)
	}

	fmt.Println("✅ 发送成功！请检查 Telegram")
}

func getAHR999Zone(ahr999 float64) string {
	if ahr999 < 0.45 {
		return "抄底区"
	} else if ahr999 < 1.20 {
		return "定投区"
	} else if ahr999 < 5.00 {
		return "持有区"
	}
	return "逃顶区"
}
