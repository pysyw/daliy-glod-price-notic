package main

import (
	"daliy-glod-price-notic/internal/cron/icbc"
	"daliy-glod-price-notic/internal/global"
	"fmt"
	"log"
)

func init() {
	global.InitHolidayManager()
}

type ILog struct{}

func (ILog) Info(msg string, keysAndValues ...interface{}) {
	log.Printf("%s, %v", msg, keysAndValues)
}

// Error logs an error condition.
func (ILog) Error(err error, msg string, keysAndValues ...interface{}) {
	panicErr := fmt.Sprintf("e%s:%s, %v", msg, err, keysAndValues)
	log.Println(panicErr)
}

// GoldInfo 金价信息结构

func main() {
	icbc.NewIcbcCron().Run()
	// ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// defer stop()
	// go func() {
	// 	c := cron.New(
	// 		cron.WithSeconds(),
	// 		cron.WithChain(cron.Recover(ILog{})))

	// 	c.AddJob("@every 1m", icbc.NewIcbcCron())
	// 	c.Start()
	// 	select {}
	// }()
	// // Listen for the interrupt signal.
	// <-ctx.Done()
	// stop()
}
