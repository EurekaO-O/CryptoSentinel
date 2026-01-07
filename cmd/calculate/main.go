// Package main 测试AHR999计算器
package main

import (
	"fmt"
	"log"

	"CryptoSentinel/internal/calculator"
	"CryptoSentinel/internal/model"
	"CryptoSentinel/internal/notifier"
	"CryptoSentinel/internal/strategy"
)

func main() {
	proxyAddr := "127.0.0.1:10809"

	fmt.Println("🧮 AHR999 自主计算测试")
	fmt.Println("========================")
	fmt.Println()

	// 创建计算器（带代理访问Binance）
	calc := calculator.NewAHR999CalculatorWithProxy(proxyAddr)

	fmt.Println("📊 正在从 Binance 获取 200 日 K线数据...")

	// 计算AHR999
	result, err := calc.Calculate()
	if err != nil {
		log.Fatalf("计算失败: %v", err)
	}

	fmt.Println()
	fmt.Println("✅ 计算完成！")
	fmt.Println("------------------------")
	fmt.Printf("📈 当前BTC价格:     $%.2f\n", result.CurrentPrice)
	fmt.Printf("💰 200日定投成本:   $%.2f\n", result.DCA200Cost)
	fmt.Printf("📅 比特币币龄:      %d 天\n", result.CoinAgeDays)
	fmt.Printf("📐 指数增长估值:    $%.2f\n", result.ExponentialValue)
	fmt.Println("------------------------")
	fmt.Printf("🎯 AHR999 指数:     %.4f\n", result.AHR999)
	fmt.Println()

	// 判断区间
	zone := getZone(result.AHR999)
	fmt.Printf("📍 当前区间: %s\n", zone)
	fmt.Println()
	fmt.Println("💡 请去 Coinglass 网页核对结果验证准确性")
	fmt.Println()

	// 询问是否发送到Telegram
	fmt.Println("📤 正在发送完整报告到 Telegram...")

	// 构建指标数据
	indicators := &model.MarketIndicators{
		CurrentPriceBTC:    result.CurrentPrice,
		CurrentPriceETH:    0, // 暂未获取ETH价格
		AHR999:             result.AHR999,
		MVRVZScore:         2.5, // Mock
		MaMultiplierState:  model.MaStateNormal,
		PiCycleCross:       false,
		EthRegressionState: model.EthRegMiddle,
		AccountLeverage:    1.0,
		Timestamp:          result.Timestamp,
		Source:             "Binance API (自主计算)",
	}

	// 执行策略
	signal := strategy.EvaluateV2(indicators)

	// 发送到Telegram
	botToken := "8530550957:AAE0popkGmCSD6hEsZaGuybOZR5w4CCEmew"
	chatID := "7150834288"
	tg := notifier.NewTelegramNotifierWithProxy(botToken, chatID, proxyAddr)

	// 添加计算详情到报告
	detailReport := fmt.Sprintf(`%s

📊 *计算详情*
- 200日定投成本: $%.2f
- 币龄: %d 天
- 指数增长估值: $%.2f
- 数据来源: Binance API (自主计算)`, signal.ReportMarkdown, result.DCA200Cost, result.CoinAgeDays, result.ExponentialValue)

	if err := tg.Send(detailReport); err != nil {
		log.Fatalf("发送失败: %v", err)
	}

	fmt.Println("✅ 发送成功！请检查 Telegram")
}

func getZone(ahr999 float64) string {
	if ahr999 < 0.45 {
		return "🟢 抄底区 (< 0.45)"
	} else if ahr999 < 1.20 {
		return "🔵 定投区 (0.45 - 1.20)"
	} else if ahr999 < 5.00 {
		return "🟡 持有区 (1.20 - 5.00)"
	}
	return "🔴 逃顶区 (> 5.00)"
}
