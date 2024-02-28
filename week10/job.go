package main

import (
	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

func InitRankingJob(redisClient redis.Cmdable, rlockClient *rlock.Client, nodeName string) *RankingJob {
	return NewRankingJob(rlockClient, redisClient, nodeName)
}

func InitJob(job *RankingJob) *cron.Cron {
	c := cron.New(cron.WithSeconds())
	c.AddJob("0/1 * * * * ?", CronJobFuncAdapter(func() error {
		return job.Run()
	}))
	return c
}
