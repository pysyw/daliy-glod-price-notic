package cfg

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
	GlobalConfig.FeiShuRobotToken = os.Getenv("FEI_SHU_ACCESS_TOKEN")

	// 解析价格阈值数组
	thresholdStr := os.Getenv("THRESHOLD_PRICE")
	if thresholdStr != "" {
		parts := strings.Split(thresholdStr, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if val, err := strconv.ParseFloat(part, 64); err == nil {
				GlobalConfig.ThresholdPrice = append(GlobalConfig.ThresholdPrice, val)
			}
		}
	}

	// 解析@用户列表
	atUserStr := os.Getenv("FEI_SHU_AT_USER")
	if atUserStr != "" {
		parts := strings.Split(atUserStr, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				GlobalConfig.FeiShuAtUser = append(GlobalConfig.FeiShuAtUser, part)
			}
		}
	}

	// 解析最大告警次数
	maxAlertStr := os.Getenv("MAX_ALERT_COUNT")
	if maxAlertStr != "" {
		if val, err := strconv.Atoi(maxAlertStr); err == nil && val > 0 {
			GlobalConfig.MaxAlertCount = val
		}
	}
	if GlobalConfig.MaxAlertCount == 0 {
		GlobalConfig.MaxAlertCount = 10 // 默认值
	}
}

var GlobalConfig = globalConfig{}

type globalConfig struct {
	FeiShuRobotToken string    // 飞书机器人 token
	ThresholdPrice   []float64 // 价格阈值数组（从大到小排序）
	FeiShuAtUser     []string  // 需要@的用户ID列表
	MaxAlertCount    int       // 每个价格区间最多发送@消息的次数
}
