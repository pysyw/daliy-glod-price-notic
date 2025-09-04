package main

import (
	"fmt"

	"daliy-glod-price-notic/cfg"
	"daliy-glod-price-notic/internal/handler"

	"github.com/CatchZeng/feishu/pkg/feishu"
	"github.com/modood/table"
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
	type ICBCTable struct {
		GoldType            string `table:"金种"`
		RealTimePrice       string `table:"实时主动积存价格"`
		UpAndDown           string `table:"涨跌"`
		LowestPrice         string `table:"最低价"`
		HightPrice          string `table:"最高价"`
		RegularDepositPrice string `table:"定期积存价"`
		UpAndDown2          string `table:"涨跌"`
		RedemptionPrice     string `table:"赎回价"`
		UpAndDown3          string `table:"涨跌"`
	}

	if len(arr) == 0 {
		return ""
	}
	hs := []ICBCTable{}
	for i := 1; i < len(arr); i++ {
		t := arr[i]
		hs = append(hs, ICBCTable{
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
	// table.Output(hs)
	// table.Table(hs)
	s := table.Table(hs)

	return s
}
