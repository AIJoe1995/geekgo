package domain

import (
	"github.com/robfig/cron/v3"
	"time"
)

type CronJob struct {
	Id    int64
	Ctime int64
	Utime int64

	Name string

	Expression string
	Cfg        string
	Executor   string
	NextTime   time.Time

	// 放弃抢占状态
	CancelFunc func() error
}

func (job CronJob) Next(t time.Time) time.Time {
	expr := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	s, _ := expr.Parse(job.Expression)
	return s.Next(t)
}
