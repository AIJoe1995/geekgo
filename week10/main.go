package main

import (
	"fmt"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
)

func main() {
	for i := 0; i < 5; i++ {
		go func() {
			redisClient := redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			})
			rlockClient := rlock.NewClient(redisClient)
			nodeName := fmt.Sprintf("node_%d", i)
			go func() {
				UpdateLoad(redisClient, nodeName)
			}()

			rankingJob := InitRankingJob(redisClient, rlockClient, nodeName)
			job := InitJob(rankingJob)
			job.Start()
			select {}
		}()
	}
	select {}

}
