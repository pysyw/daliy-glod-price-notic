package main

import (
	"context"
	"daliy-glod-price-notic/internal/cron/icbc"
	"daliy-glod-price-notic/internal/global"
	"daliy-glod-price-notic/internal/server"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"
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
	// 启动的时候发送一次
	icbc.NewIcbcCron().Run()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 启动定时任务
	go func() {
		c := cron.New(
			cron.WithSeconds(),
			cron.WithChain(cron.Recover(ILog{})))

		c.AddJob("@every 1m", icbc.NewIcbcCron())
		c.Start()
		select {}
	}()

	// 启动 HTTP 服务器
	go func() {
		srv := server.NewServer()
		if err := srv.Start(); err != nil {
			log.Printf("HTTP 服务器启动失败: %v\n", err)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()
	stop()
	log.Println("程序已退出")
}
