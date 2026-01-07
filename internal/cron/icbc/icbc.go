package icbc

import (
	"daliy-glod-price-notic/cfg"
	"daliy-glod-price-notic/internal/global"
	"daliy-glod-price-notic/internal/handler"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CatchZeng/feishu/pkg/feishu"
)

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
	feishuClient := feishu.NewClient(cfg.GlobalConfig.FeiShuRobotToken, "")
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

	cardJSON := buildFeishuCard(goldList)
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

// buildFeishuCard 构建飞书卡片消息
func buildFeishuCard(goldList []GoldInfo) string {
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
		title = fmt.Sprintf("%s:%s", title, goldList[0].RealTimePrice)
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
