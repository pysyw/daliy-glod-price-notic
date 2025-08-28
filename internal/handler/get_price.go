package handler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
)

func GetICBCGoldPrice() (result [][]string, err error) {
	result = [][]string{}
	targetUrl := "https://mybank.icbc.com.cn/icbc/newperbank/perbank3/gold/goldaccrual_query_out.jsp"
	resp, err := req.C().R().Get(targetUrl)
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("请求工商实时金价失败, 状态码:%d", resp.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	doc.Find("#resultTable>tbody").Children().Each(func(i int, tr *goquery.Selection) {
		arr := []string{}
		tr.Children().Each(func(i int, s *goquery.Selection) {
			pattern := regexp.MustCompile(`\s+`)
			res := pattern.ReplaceAllString(s.Text(), " ")
			arr = append(arr, strings.TrimSpace(res))
		})
		result = append(result, arr)
	})

	return result, nil
}
