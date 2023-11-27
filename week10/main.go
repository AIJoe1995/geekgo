package main

import (
	"geekgo/week10/job"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	load := job.NewLoad(redisClient)
	redisLockClient := rlock.NewClient(redisClient)
	job := job.NewRankingJob(redisLockClient, load)
	job.Run()
}
