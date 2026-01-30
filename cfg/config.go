package cfg

import (
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()

	// 初始化运行时配置（从环境变量加载默认值）
	runtimeConfig = &RuntimeConfig{}
	runtimeConfig.Load()

	// 保留 GlobalConfig 用于向后兼容（已废弃，建议使用 GetRuntimeConfig）
	GlobalConfig.FeiShuRobotToken = runtimeConfig.GetFeiShuToken()
	GlobalConfig.ThresholdPrice = runtimeConfig.GetThresholdPrice()
	GlobalConfig.FeiShuAtUser = runtimeConfig.GetFeiShuAtUser()
	GlobalConfig.MaxAlertCount = runtimeConfig.GetMaxAlertCount()
}

// RuntimeConfig 运行时配置（线程安全）
type RuntimeConfig struct {
	mu               sync.RWMutex
	feiShuToken      string
	thresholdPrice   []float64
	feiShuAtUser     []string
	maxAlertCount    int
}

var runtimeConfig *RuntimeConfig

// GetRuntimeConfig 获取全局运行时配置实例
func GetRuntimeConfig() *RuntimeConfig {
	return runtimeConfig
}

// Load 从环境变量加载配置
func (c *RuntimeConfig) Load() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.feiShuToken = os.Getenv("FEI_SHU_ACCESS_TOKEN")

	// 解析价格阈值数组
	thresholdStr := os.Getenv("THRESHOLD_PRICE")
	c.thresholdPrice = []float64{}
	if thresholdStr != "" {
		parts := strings.Split(thresholdStr, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if val, err := strconv.ParseFloat(part, 64); err == nil {
				c.thresholdPrice = append(c.thresholdPrice, val)
			}
		}
	}

	// 解析@用户列表
	atUserStr := os.Getenv("FEI_SHU_AT_USER")
	c.feiShuAtUser = []string{}
	if atUserStr != "" {
		parts := strings.Split(atUserStr, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				c.feiShuAtUser = append(c.feiShuAtUser, part)
			}
		}
	}

	// 解析最大告警次数
	maxAlertStr := os.Getenv("MAX_ALERT_COUNT")
	c.maxAlertCount = 10 // 默认值
	if maxAlertStr != "" {
		if val, err := strconv.Atoi(maxAlertStr); err == nil && val > 0 {
			c.maxAlertCount = val
		}
	}
}

// Getter 方法（线程安全）

func (c *RuntimeConfig) GetFeiShuToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.feiShuToken
}

func (c *RuntimeConfig) GetThresholdPrice() []float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]float64, len(c.thresholdPrice))
	copy(result, c.thresholdPrice)
	return result
}

func (c *RuntimeConfig) GetFeiShuAtUser() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, len(c.feiShuAtUser))
	copy(result, c.feiShuAtUser)
	return result
}

func (c *RuntimeConfig) GetMaxAlertCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxAlertCount
}

// Setter 方法（线程安全）

func (c *RuntimeConfig) SetThresholdPrice(prices []float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.thresholdPrice = make([]float64, len(prices))
	copy(c.thresholdPrice, prices)
}

func (c *RuntimeConfig) SetFeiShuAtUser(users []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.feiShuAtUser = make([]string, len(users))
	copy(c.feiShuAtUser, users)
}

func (c *RuntimeConfig) SetMaxAlertCount(count int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if count > 0 {
		c.maxAlertCount = count
	}
}

// GetAllConfig 获取所有配置（用于展示）
func (c *RuntimeConfig) GetAllConfig() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"threshold_price": c.thresholdPrice,
		"feishu_at_user":  c.feiShuAtUser,
		"max_alert_count": c.maxAlertCount,
	}
}

// GlobalConfig 全局配置（已废弃，保留用于向后兼容）
var GlobalConfig = globalConfig{}

type globalConfig struct {
	FeiShuRobotToken string    // 飞书机器人 token
	ThresholdPrice   []float64 // 价格阈值数组（从大到小排序）
	FeiShuAtUser     []string  // 需要@的用户ID列表
	MaxAlertCount    int       // 每个价格区间最多发送@消息的次数
}
