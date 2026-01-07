// Package config 提供配置文件读取和管理功能
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 应用程序配置结构体
type Config struct {
	Telegram  TelegramConfig  `yaml:"telegram"`
	DataSource DataSourceConfig `yaml:"data_source"`
	Schedule  ScheduleConfig  `yaml:"schedule"`
}

// TelegramConfig Telegram机器人配置
type TelegramConfig struct {
	// BotToken Telegram Bot API Token
	// 可通过环境变量 TELEGRAM_BOT_TOKEN 覆盖
	BotToken string `yaml:"bot_token"`
	// ChatID 目标聊天ID
	// 可通过环境变量 TELEGRAM_CHAT_ID 覆盖
	ChatID string `yaml:"chat_id"`
}

// DataSourceConfig 数据源配置
type DataSourceConfig struct {
	// AHR999URL AHR999指数数据接口URL
	// 可通过环境变量 AHR999_URL 覆盖
	AHR999URL string `yaml:"ahr999_url"`
}

// ScheduleConfig 定时任务配置
type ScheduleConfig struct {
	// CronSpec Cron表达式，定义执行时间
	// 可通过环境变量 CRON_SPEC 覆盖
	CronSpec string `yaml:"cron_spec"`
}

// Load 从指定路径加载配置文件
// 环境变量优先级高于配置文件
func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	// 读取配置文件
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("获取配置文件绝对路径失败: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 环境变量覆盖配置文件
	cfg.overrideWithEnv()

	// 验证必要配置
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return cfg, nil
}

// overrideWithEnv 使用环境变量覆盖配置
func (c *Config) overrideWithEnv() {
	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		c.Telegram.BotToken = token
	}
	if chatID := os.Getenv("TELEGRAM_CHAT_ID"); chatID != "" {
		c.Telegram.ChatID = chatID
	}
	if url := os.Getenv("AHR999_URL"); url != "" {
		c.DataSource.AHR999URL = url
	}
	if cronSpec := os.Getenv("CRON_SPEC"); cronSpec != "" {
		c.Schedule.CronSpec = cronSpec
	}
}

// validate 验证配置完整性
func (c *Config) validate() error {
	if c.Telegram.BotToken == "" {
		return fmt.Errorf("Telegram Bot Token 未配置")
	}
	if c.Telegram.ChatID == "" {
		return fmt.Errorf("Telegram Chat ID 未配置")
	}
	if c.DataSource.AHR999URL == "" {
		return fmt.Errorf("AHR999 数据源 URL 未配置")
	}
	if c.Schedule.CronSpec == "" {
		return fmt.Errorf("Cron 表达式未配置")
	}
	return nil
}
