package main

import (
	"fmt"

	"daliy-glod-price-notic/cfg"
	"daliy-glod-price-notic/internal/handler"

	"github.com/CatchZeng/feishu/pkg/feishu"
)

func main() {
	feishuClient := feishu.NewClient(cfg.GlobalConfig.FeiShuRobotToken, "")
	result, err := handler.GetICBCGoldPrice()
	if err != nil {
		feishuClient.Send(feishu.NewTextMessage().SetText(fmt.Sprintf("请求工商实时金价失败:%s", err.Error())))
		return
	}
	fmt.Println(formatStr(result))
	// feishuClient.Send(feishu.ht)
	// feishuClient.Send(feishu.NewTextMessage().SetText(formatStr(result)))
}

// TODO:格式化输出
func formatStr(arr [][]string) string {
	return ""
}
