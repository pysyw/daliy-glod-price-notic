package global

import (
	"time"
)

type icbc struct{}

var Icbc = &icbc{}

func (*icbc) IsWorkTime() bool {
	now := time.Now()
	// 9:10-22:30
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 9, 10, 0, 0, time.Local)
	endTime := time.Date(now.Year(), now.Month(), now.Day(), 22, 30, 0, 0, time.Local)

	return now.Unix() >= startTime.Unix() && now.Unix() <= endTime.Unix()
}
