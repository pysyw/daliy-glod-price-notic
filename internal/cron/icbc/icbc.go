package icbc

import (
	"daliy-glod-price-notic/cfg"
	"daliy-glod-price-notic/internal/global"
	"daliy-glod-price-notic/internal/handler"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/CatchZeng/feishu/pkg/feishu"
	"github.com/patrickmn/go-cache"
)

// alertCache 用于控制@消息发送次数，key为价格区间标识，value为发送次数
// 过期时间1小时，自动清理间隔10分钟
var alertCache = cache.New(1*time.Hour, 10*time.Minute)

type icbcCron struct{}

func NewIcbcCron() *icbcCron {
	return &icbcCron{}
}

func (s *icbcCron) Run() {
	// 如果非工作时间 也不推送
	if !global.Icbc.IsWorkTime() {
		return
	}
	// 如果是假期就不需要推送了
	if global.IsHoliday(time.Now()) {
		return
	}
	feishuClient := feishu.NewClient(cfg.GetRuntimeConfig().GetFeiShuToken(), "")
	sendGoldPrice(feishuClient)
}

// sendGoldPrice 获取并发送金价信息
func sendGoldPrice(client *feishu.Client) {
	// 添加 panic 恢复，避免单次执行失败导致整个程序崩溃
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[%s] sendGoldPrice 发生 panic: %v\n", time.Now().Format("2006-01-02 15:04:05"), r)
		}
	}()

	fmt.Printf("[%s] 开始获取金价...\n", time.Now().Format("2006-01-02 15:04:05"))

	result, err := handler.GetICBCGoldPrice()
	if err != nil {
		client.Send(feishu.NewTextMessage().SetText(fmt.Sprintf("请求工商实时金价失败: %s", err.Error())))
		fmt.Printf("[%s] 获取金价失败: %s\n", time.Now().Format("2006-01-02 15:04:05"), err.Error())
		return
	}

	goldList := parseGoldData(result)
	if len(goldList) == 0 {
		client.Send(feishu.NewTextMessage().SetText("未获取到金价数据"))
		fmt.Printf("[%s] 未获取到金价数据\n", time.Now().Format("2006-01-02 15:04:05"))
		return
	}

	// 检查是否需要发送@消息
	currentPrice := ""
	if len(goldList) > 0 {
		currentPrice = goldList[0].RealTimePrice
	}
	shouldAlert, alertInfo := checkPriceAlert(currentPrice, cfg.GetRuntimeConfig())

	// 构建卡片消息（包含@信息）
	cardJSON := buildFeishuCard(goldList, shouldAlert, alertInfo)
	msg := feishu.NewInteractiveMessage().SetCard(cardJSON)
	client.Send(msg)

	fmt.Printf("[%s] 金价推送成功\n", time.Now().Format("2006-01-02 15:04:05"))
}

// parseGoldData 解析金价数据
func parseGoldData(arr [][]string) []GoldInfo {
	var result []GoldInfo
	for i := 1; i < len(arr); i++ {
		t := arr[i]
		if len(t) < 9 {
			continue
		}
		result = append(result, GoldInfo{
			GoldType:            t[0],
			RealTimePrice:       t[1],
			UpAndDown:           t[2],
			LowestPrice:         t[3],
			HightPrice:          t[4],
			RegularDepositPrice: t[5],
			UpAndDown2:          t[6],
			RedemptionPrice:     t[7],
			UpAndDown3:          t[8],
		})
	}
	return result
}

// formatTrend 格式化涨跌趋势显示
func formatTrend(trend string) string {
	switch trend {
	case "涨":
		return "🔺"
	case "跌":
		return "🔻"
	default:
		return "➖"
	}
}

type GoldInfo struct {
	GoldType            string // 金种
	RealTimePrice       string // 实时主动积存价格
	UpAndDown           string // 涨跌
	LowestPrice         string // 最低价
	HightPrice          string // 最高价
	RegularDepositPrice string // 定期积存价
	UpAndDown2          string // 涨跌
	RedemptionPrice     string // 赎回价
	UpAndDown3          string // 涨跌
}

// AlertInfo 告警信息
type AlertInfo struct {
	IntervalIndex int      // 价格区间索引
	Lower         float64  // 区间下限
	Upper         float64  // 区间上限
	MaxAlertCount int      // 该区间最大告警次数
	UserIDs       []string // 需要@的用户ID列表
}

// buildFeishuCard 构建飞书卡片消息
func buildFeishuCard(goldList []GoldInfo, shouldAlert bool, alertInfo *AlertInfo) string {
	var elements []map[string]interface{}

	// 添加分割线
	elements = append(elements, map[string]interface{}{
		"tag": "hr",
	})

	for _, gold := range goldList {
		// 金种标题
		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**🏅 %s**", gold.GoldType),
			},
		})

		// 构建价格信息（使用列布局）
		var fields []map[string]interface{}

		// 实时积存价
		fields = append(fields, map[string]interface{}{
			"is_short": true,
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**实时积存价**\n¥ %s %s", gold.RealTimePrice, formatTrend(gold.UpAndDown)),
			},
		})

		// 定期积存价
		fields = append(fields, map[string]interface{}{
			"is_short": true,
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**定期积存价**\n¥ %s %s", gold.RegularDepositPrice, formatTrend(gold.UpAndDown2)),
			},
		})

		// 赎回价
		fields = append(fields, map[string]interface{}{
			"is_short": true,
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**赎回价**\n¥ %s %s", gold.RedemptionPrice, formatTrend(gold.UpAndDown3)),
			},
		})

		// 今日区间
		if gold.LowestPrice != "----" && gold.HightPrice != "----" {
			fields = append(fields, map[string]interface{}{
				"is_short": true,
				"text": map[string]interface{}{
					"tag":     "lark_md",
					"content": fmt.Sprintf("**今日区间**\n¥ %s ~ %s", gold.LowestPrice, gold.HightPrice),
				},
			})
		}

		elements = append(elements, map[string]interface{}{
			"tag":    "div",
			"fields": fields,
		})

		// 添加分割线
		elements = append(elements, map[string]interface{}{
			"tag": "hr",
		})
	}

	// 如果需要告警，添加@用户
	if shouldAlert && alertInfo != nil {
		// 构建@用户的文本
		atText := ""
		for _, userID := range alertInfo.UserIDs {
			atText += fmt.Sprintf("<at id=\"%s\"></at> ", userID)
		}

		elements = append(elements, map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": atText,
			},
		})
	}

	// 添加当前配置展示和修改按钮
	runtimeCfg := cfg.GetRuntimeConfig()
	intervals := runtimeCfg.GetPriceIntervals()
	atUsers := runtimeCfg.GetFeiShuAtUser()

	// 格式化配置显示
	configStr := "未配置"
	if len(intervals) > 0 {
		var parts []string
		for _, interval := range intervals {
			parts = append(parts, fmt.Sprintf("[%.2f, %.2f) 最多%d次", interval.Lower, interval.Upper, interval.MaxAlertCount))
		}
		configStr = strings.Join(parts, "\n")
	}

	// 格式化@用户显示
	atUserStr := "未配置"
	if len(atUsers) > 0 {
		atUserStr = fmt.Sprintf("%d 个用户", len(atUsers))
	}

	// 配置信息展示
	elements = append(elements, map[string]interface{}{
		"tag": "div",
		"text": map[string]interface{}{
			"tag":     "lark_md",
			"content": fmt.Sprintf("**⚙️ 当前推送配置**\n价格区间：\n%s\n@用户数：%s", configStr, atUserStr),
		},
	})

	// 修改配置按钮
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // 默认值
	}
	elements = append(elements, map[string]interface{}{
		"tag": "action",
		"actions": []map[string]interface{}{
			{
				"tag": "button",
				"text": map[string]interface{}{
					"tag":     "plain_text",
					"content": "🔧 修改配置",
				},
				"type": "primary",
				"url":  fmt.Sprintf("%s/config", baseURL),
			},
		},
	})

	// 添加分割线
	elements = append(elements, map[string]interface{}{
		"tag": "hr",
	})

	// 添加备注
	elements = append(elements, map[string]interface{}{
		"tag": "note",
		"elements": []map[string]interface{}{
			{
				"tag":     "plain_text",
				"content": fmt.Sprintf("数据来源：工商银行 | 更新时间：%s", time.Now().Format("2006-01-02 15:04:05")),
			},
		},
	})

	title := "📊 工行实时价格"
	if len(goldList) > 0 {
		currentPrice := goldList[0].RealTimePrice
		title = fmt.Sprintf("%s:%s", title, currentPrice)
	}

	card := map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
			"enable_forward":   true,
		},
		"header": map[string]interface{}{
			"title": map[string]interface{}{
				"tag":     "plain_text",
				"content": title,
			},
			"template": "gold",
		},
		"elements": elements,
	}

	cardBytes, _ := json.Marshal(card)
	return string(cardBytes)
}

// checkPriceAlert 检查价格是否需要告警
// 返回值：是否需要告警，告警信息
func checkPriceAlert(priceStr string, runtimeCfg *cfg.RuntimeConfig) (bool, *AlertInfo) {
	// 检查配置是否完整
	if runtimeCfg == nil {
		return false, nil
	}

	atUsers := runtimeCfg.GetFeiShuAtUser()
	priceIntervals := runtimeCfg.GetPriceIntervals()

	if len(priceIntervals) == 0 || len(atUsers) == 0 {
		return false, nil
	}

	// 解析当前价格
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		fmt.Printf("[%s] 解析价格失败: %s\n", time.Now().Format("2006-01-02 15:04:05"), err.Error())
		return false, nil
	}

	// 查找价格所在的区间
	for i, interval := range priceIntervals {
		if price >= interval.Lower && price < interval.Upper {
			// 检查该区间是否应该发送@消息
			if !shouldSendAlert(i, interval.MaxAlertCount) {
				fmt.Printf("[%s] 价格区间 %d [%.2f, %.2f) 已达到最大告警次数 %d，不再发送@消息\n",
					time.Now().Format("2006-01-02 15:04:05"), i, interval.Lower, interval.Upper, interval.MaxAlertCount)
				return false, nil
			}

			// 增加该区间的告警计数
			incrementAlertCount(i)

			// 构建告警信息
			alertInfo := &AlertInfo{
				IntervalIndex: i,
				Lower:         interval.Lower,
				Upper:         interval.Upper,
				MaxAlertCount: interval.MaxAlertCount,
				UserIDs:       atUsers,
			}

			fmt.Printf("[%s] 价格告警：价格区间 %d [%.2f, %.2f)，当前价格 %.2f，最大告警次数 %d\n",
				time.Now().Format("2006-01-02 15:04:05"), i, interval.Lower, interval.Upper, price, interval.MaxAlertCount)

			return true, alertInfo
		}
	}

	// 价格不在任何配置的区间内
	return false, nil
}

// shouldSendAlert 判断是否应该发送@消息
func shouldSendAlert(intervalIndex int, maxCount int) bool {
	cacheKey := fmt.Sprintf("alert_interval_%d", intervalIndex)
	countVal, found := alertCache.Get(cacheKey)
	if !found {
		return true
	}

	count, ok := countVal.(int)
	if !ok {
		return true
	}

	return count < maxCount
}

// incrementAlertCount 增加指定区间的告警计数
func incrementAlertCount(intervalIndex int) {
	cacheKey := fmt.Sprintf("alert_interval_%d", intervalIndex)
	countVal, found := alertCache.Get(cacheKey)

	count := 0
	if found {
		if c, ok := countVal.(int); ok {
			count = c
		}
	}

	count++
	alertCache.Set(cacheKey, count, cache.DefaultExpiration)
}
