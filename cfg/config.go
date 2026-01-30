package cfg

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()

	// 初始化运行时配置
	runtimeConfig = &RuntimeConfig{}
	runtimeConfig.Load()
}

// PriceInterval 价格区间配置
type PriceInterval struct {
	Lower         float64 // 区间下限（包含）
	Upper         float64 // 区间上限（不包含）
	MaxAlertCount int     // 该区间最大告警次数
}

// RuntimeConfig 运行时配置（线程安全）
type RuntimeConfig struct {
	mu             sync.RWMutex
	feiShuToken    string
	priceIntervals []PriceInterval // 价格区间列表
	feiShuAtUser   []string
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

	// 解析价格区间配置
	intervalsStr := os.Getenv("PRICE_INTERVALS")
	if intervalsStr != "" {
		c.priceIntervals = parsePriceIntervals(intervalsStr)
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
}

// parsePriceIntervals 解析价格区间配置字符串
// 格式：下限-上限:次数,下限-上限:次数
// 示例：1045-1047:5,1047-1051:10
func parsePriceIntervals(str string) []PriceInterval {
	var intervals []PriceInterval
	parts := strings.Split(str, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 分离区间和告警次数（格式：下限-上限:次数）
		segments := strings.Split(part, ":")
		if len(segments) != 2 {
			fmt.Printf("警告：跳过无效的价格区间配置 '%s'（格式应为 下限-上限:次数）\n", part)
			continue
		}

		rangeStr := strings.TrimSpace(segments[0])
		countStr := strings.TrimSpace(segments[1])

		// 解析区间上下限
		bounds := strings.Split(rangeStr, "-")
		if len(bounds) != 2 {
			fmt.Printf("警告：跳过无效的价格区间 '%s'（格式应为 下限-上限）\n", rangeStr)
			continue
		}

		lower, err1 := strconv.ParseFloat(strings.TrimSpace(bounds[0]), 64)
		upper, err2 := strconv.ParseFloat(strings.TrimSpace(bounds[1]), 64)
		count, err3 := strconv.Atoi(countStr)

		if err1 != nil || err2 != nil || err3 != nil {
			fmt.Printf("警告：跳过无效的价格区间配置 '%s'（数字解析失败）\n", part)
			continue
		}

		if lower >= upper {
			fmt.Printf("警告：跳过无效的价格区间 '%s'（下限必须小于上限）\n", part)
			continue
		}

		if count <= 0 {
			fmt.Printf("警告：跳过无效的价格区间 '%s'（告警次数必须大于0）\n", part)
			continue
		}

		intervals = append(intervals, PriceInterval{
			Lower:         lower,
			Upper:         upper,
			MaxAlertCount: count,
		})
	}

	// 按下限从小到大排序
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i].Lower < intervals[j].Lower
	})

	return intervals
}

// Getter 方法（线程安全）

func (c *RuntimeConfig) GetFeiShuToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.feiShuToken
}

// GetPriceIntervals 获取价格区间配置
func (c *RuntimeConfig) GetPriceIntervals() []PriceInterval {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]PriceInterval, len(c.priceIntervals))
	copy(result, c.priceIntervals)
	return result
}

func (c *RuntimeConfig) GetFeiShuAtUser() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, len(c.feiShuAtUser))
	copy(result, c.feiShuAtUser)
	return result
}

// Setter 方法（线程安全）

// SetPriceIntervals 设置价格区间配置
func (c *RuntimeConfig) SetPriceIntervals(intervals []PriceInterval) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.priceIntervals = make([]PriceInterval, len(intervals))
	copy(c.priceIntervals, intervals)
}

func (c *RuntimeConfig) SetFeiShuAtUser(users []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.feiShuAtUser = make([]string, len(users))
	copy(c.feiShuAtUser, users)
}

// GetAllConfig 获取所有配置（用于展示）
func (c *RuntimeConfig) GetAllConfig() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]any{
		"price_intervals": c.priceIntervals,
		"feishu_at_user":  c.feiShuAtUser,
	}
}
