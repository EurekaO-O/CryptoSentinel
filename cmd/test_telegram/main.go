// Package main 测试Telegram发送功能
package main

import (
	"fmt"
	"log"
	"time"

	"CryptoSentinel/internal/model"
	"CryptoSentinel/internal/notifier"
	"CryptoSentinel/internal/strategy"
)

func main() {
	// 直接使用配置好的Token和ChatID
	botToken := "8530550957:AAE0popkGmCSD6hEsZaGuybOZR5w4CCEmew"
	chatID := "7150834288"

	fmt.Println("正在发送测试消息到 Telegram...")

	// 创建带代理的通知器
	tg := notifier.NewTelegramNotifierWithProxy(botToken, chatID, "127.0.0.1:10809")

	// 模拟当前市场数据
	indicators := &model.MarketIndicators{
		CurrentPriceBTC:    96500,
		CurrentPriceETH:    3580,
		AHR999:             0.72,
		MVRVZScore:         2.3,
		MaMultiplierState:  model.MaStateNormal,
		PiCycleCross:       false,
		EthRegressionState: model.EthRegMiddle,
		AccountLeverage:    1.0,
		Timestamp:          time.Now(),
	}

	// 执行策略
	signal := strategy.EvaluateV2(indicators)

	// 发送报告
	err := tg.Send(signal.ReportMarkdown)
	if err != nil {
		log.Fatalf("发送失败: %v", err)
	}

	fmt.Println("✅ 发送成功！请检查 Telegram")
}
