package ioc

import (
	"context"
	"fmt"
	"geekgo/week11/domain"
	"geekgo/week11/job"
	"geekgo/week11/service"
	"time"
)

func InitScheduler(local *job.LocalFuncExecutor, svc service.CronJobService) *job.Scheduler {
	res := job.NewScheduler(svc)
	res.RegisterExecutor(local)
	return res
}

func InitLocalFuncExecutor() *job.LocalFuncExecutor {
	res := job.NewLocalFuncExecutor()
	res.RegisterFunc("ranking", func(ctx context.Context, j domain.CronJob) error {
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		fmt.Println("ranking", time.Now().String())
		return nil
	})
	return res
}
