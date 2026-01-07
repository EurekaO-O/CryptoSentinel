// Package main CryptoSentinel 加密货币定投监控机器人入口
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"CryptoSentinel/internal/collector"
	"CryptoSentinel/internal/config"
	"CryptoSentinel/internal/notifier"
	"CryptoSentinel/internal/strategy"

	"github.com/robfig/cron/v3"
)

// 默认杠杆率（可通过环境变量 LEVERAGE 设置）
const defaultLeverage = 1.0

func main() {
	// 初始化日志
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("CryptoSentinel 启动中...")

	// 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	log.Println("配置加载成功")

	// 初始化各模块
	dataCollector := collector.NewCollector(cfg.DataSource.AHR999URL)
	telegramNotifier := notifier.NewTelegramNotifier(cfg.Telegram.BotToken, cfg.Telegram.ChatID)

	// 获取杠杆率配置
	leverage := defaultLeverage
	if envLeverage := os.Getenv("LEVERAGE"); envLeverage != "" {
		if _, err := fmt.Sscanf(envLeverage, "%f", &leverage); err != nil {
			log.Printf("解析环境变量 LEVERAGE 失败，使用默认值: %v", err)
		}
	}

	// 定义定时任务执行函数
	taskFunc := func() {
		log.Println("执行定时任务...")

		// 1. 采集数据
		marketData, err := dataCollector.FetchAHR999()
		if err != nil {
			errMsg := fmt.Sprintf("获取AHR999数据失败: %v", err)
			log.Println(errMsg)
			if sendErr := telegramNotifier.SendWithRetry(errMsg, 3); sendErr != nil {
				log.Printf("发送错误通知失败: %v", sendErr)
			}
			return
		}
		log.Printf("获取到AHR999数据: %.4f", marketData.AHR999)

		// 2. 执行策略分析
		decision := strategy.Evaluate(marketData.AHR999, leverage)
		log.Printf("策略决策: %s", decision.Action)

		// 3. 生成通知消息
		message := formatMessage(marketData, decision, leverage)

		// 4. 发送通知
		if err := telegramNotifier.SendWithRetry(message, 3); err != nil {
			log.Printf("发送Telegram通知失败: %v", err)
			return
		}
		log.Println("通知发送成功")
	}

	// 初始化Cron调度器
	c := cron.New(cron.WithSeconds())
	_, err = c.AddFunc(cfg.Schedule.CronSpec, taskFunc)
	if err != nil {
		log.Fatalf("添加定时任务失败: %v", err)
	}

	// 启动调度器
	c.Start()
	log.Printf("定时任务已启动，Cron表达式: %s", cfg.Schedule.CronSpec)

	// 启动时立即执行一次（可选）
	if os.Getenv("RUN_ON_START") == "true" {
		log.Println("启动时执行一次任务...")
		taskFunc()
	}

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("收到退出信号，正在关闭...")
	c.Stop()
	log.Println("CryptoSentinel 已关闭")
}

// formatMessage 格式化Telegram通知消息
func formatMessage(data *collector.MarketData, decision *strategy.Decision, leverage float64) string {
	actionEmoji := getActionEmoji(decision.Action)
	return fmt.Sprintf(`*CryptoSentinel 定投提醒*

%s *%s*

*AHR999指数:* %.4f
*当前杠杆率:* %.2f
*资金倍率:* %.1fx

*分析:*
%s

_数据来源: %s_
_时间: %s_`,
		actionEmoji,
		decision.Action,
		data.AHR999,
		leverage,
		decision.AmountFactor,
		decision.Message,
		data.Source,
		data.Timestamp.Format("2006-01-02 15:04:05"),
	)
}

// getActionEmoji 根据动作返回对应的emoji
func getActionEmoji(action strategy.Action) string {
	switch action {
	case strategy.ActionWarning:
		return "🚨"
	case strategy.ActionGreedyBuy:
		return "🟢"
	case strategy.ActionDCA:
		return "🔵"
	case strategy.ActionHold:
		return "🟡"
	case strategy.ActionSell:
		return "🔴"
	default:
		return "📊"
	}
}
