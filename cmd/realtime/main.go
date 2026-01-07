// Package main 实时数据测试 - 模拟当前市场数据并发送报告
package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"CryptoSentinel/internal/model"
	"CryptoSentinel/internal/notifier"
	"CryptoSentinel/internal/strategy"
)

func main() {
	botToken := "8530550957:AAE0popkGmCSD6hEsZaGuybOZR5w4CCEmew"
	chatID := "7150834288"
	proxyAddr := "127.0.0.1:10809"

	fmt.Println("📊 生成模拟市场数据...")

	// 模拟真实市场数据（API暂不可用，使用模拟值）
	// 真实AHR999当前大约在0.5-0.8区间
	rand.Seed(time.Now().UnixNano())
	ahr999 := 0.65 + rand.Float64()*0.15 // 0.65-0.80 随机

	indicators := &model.MarketIndicators{
		CurrentPriceBTC:    96500 + rand.Float64()*1000,
		CurrentPriceETH:    3550 + rand.Float64()*100,
		AHR999:             ahr999,
		MVRVZScore:         2.1 + rand.Float64()*0.5,
		MaMultiplierState:  model.MaStateNormal,
		PiCycleCross:       false,
		EthRegressionState: model.EthRegMiddle,
		AccountLeverage:    1.0,
		Timestamp:          time.Now(),
		Source:             "模拟数据",
	}

	fmt.Printf("✅ BTC价格: $%.0f\n", indicators.CurrentPriceBTC)
	fmt.Printf("✅ ETH价格: $%.0f\n", indicators.CurrentPriceETH)
	fmt.Printf("✅ AHR999: %.4f\n", indicators.AHR999)
	fmt.Printf("✅ MVRV-Z: %.2f\n", indicators.MVRVZScore)

	// 执行策略分析
	signal := strategy.EvaluateV2(indicators)
	fmt.Printf("📈 策略建议: %s\n", signal.ActionBTC)

	// 发送到 Telegram
	fmt.Println("📤 正在发送到 Telegram...")
	tg := notifier.NewTelegramNotifierWithProxy(botToken, chatID, proxyAddr)

	if err := tg.Send(signal.ReportMarkdown); err != nil {
		log.Fatalf("发送失败: %v", err)
	}

	fmt.Println("✅ 发送成功！请检查 Telegram")
}
