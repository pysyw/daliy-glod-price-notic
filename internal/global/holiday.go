package global

import (
	"sync"
	"time"

	"github.com/bastengao/chinese-holidays-go/holidays"
)

type GlobalHolidayManager struct {
	queryer holidays.Queryer
	mu      sync.RWMutex
}

var holidayMgr *GlobalHolidayManager

// InitHolidayManager 初始化（在 main 或 init 中调用一次）
func InitHolidayManager() error {
	q, err := holidays.NewCacheQueryer()
	if err != nil {
		return err
	}

	holidayMgr = &GlobalHolidayManager{
		queryer: q,
		mu:      sync.RWMutex{},
	}

	return nil
}

// IsHoliday 安全地对外提供查询方法
func IsHoliday(t time.Time) bool {
	holidayMgr.mu.RLock() // 加读锁，并发安全
	defer holidayMgr.mu.RUnlock()
	result, _ := holidayMgr.queryer.IsHoliday(t)
	return result
}

func IsWorkingday(t time.Time) bool {
	holidayMgr.mu.RLock() // 加读锁，并发安全
	defer holidayMgr.mu.RUnlock()
	result, _ := holidayMgr.queryer.IsWorkingday(t)
	return result
}
