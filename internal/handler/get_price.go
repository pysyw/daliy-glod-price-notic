package handler

import (
	"fmt"
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
			trChildTag := s.Get(0).Data
			if trChildTag == "th" {
				arr = append(arr, strings.TrimSpace(s.Text()))
			} else if trChildTag == "td" {
				if len(s.Children().Nodes) == 0 {
					arr = append(arr, strings.TrimSpace(s.Text()))
				} else {
					child := s.Children().Nodes[0]
					if child.Data == "img" {
						alt, _ := s.Children().Attr("alt")
						arr = append(arr, alt)
					}
				}
			}
		})
		result = append(result, arr)
	})

	return result, nil
}
