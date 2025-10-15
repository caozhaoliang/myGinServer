package cron

import (
	"fmt"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

func TestCron(t *testing.T) {
	c := cron.New()

	// 每秒执行一次
	_, err := c.AddFunc("@every 1s", func() {
		fmt.Printf("A每秒任务执行: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	})
	if err != nil {
		fmt.Println("添加任务失败:", err)
		return
	}

	// 每5秒执行一次
	_, err = c.AddFunc("@every 5s", func() {
		fmt.Printf("B每5秒任务执行: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	})
	if err != nil {
		fmt.Println("添加任务失败:", err)
		return
	}

	c.Start()
	defer c.Stop()

	// 让程序运行一段时间
	time.Sleep(30 * time.Second)
}

func TestCronTime(t *testing.T) {
	c := cron.New()

	c.AddFunc("30 * * * *", func() {
		fmt.Println("Every hour on the half hour")
	})

	c.AddFunc("30 3-6,20-23 * * *", func() {
		fmt.Println("On the half hour of 3-6am, 8-11pm")
	})

	c.AddFunc("0 0 1 1 *", func() {
		fmt.Println("Jun 1 every year")
	})

	c.Start()

	for {
		time.Sleep(time.Second)
	}
}

func TestSimpleCron(t *testing.T) {
	c := cron.New()

	c.AddFunc("@hourly", func() {
		fmt.Println("Every hour")
	})

	c.AddFunc("@daily", func() {
		fmt.Println("Every day on midnight")
	})

	c.AddFunc("@weekly", func() {
		fmt.Println("Every week")
	})

	c.Start()

	for {
		time.Sleep(time.Second)
	}
}

func TestCronWithLocation(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		fmt.Println("加载时区失败:", err)
		return
	}

	c := cron.New(cron.WithLocation(loc))

	c.AddFunc("@every 1s", func() {
		fmt.Printf("纽约时区每秒任务执行: %s\n", time.Now().In(loc).Format("2006-01-02 15:04:05"))
	})

	c.AddFunc("CRON_TZ=Asia/Tokyo 0 6 * * ?", func() {
		fmt.Println("Every 6 o'clock at Tokyo")
	})

	c.Start()
	select {}
}

func TestCronNewParser(t *testing.T) {

	c := cron.New(cron.WithSeconds())
	_, err := c.AddFunc("* * * * * *", func() {
		fmt.Println("every 1 second")
	})
	if err != nil {
		t.Error(err)
	}
	c.Start()
	select {}
}
