package main

import (
	"context"
	"fmt"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

type RankingJob struct {
	client  *rlock.Client
	key     string
	timeout time.Duration
	job     func(ctx context.Context) error

	localLock   sync.Mutex
	lock        *rlock.Lock
	redisClient redis.Cmdable
	nodeName    string
	threshhold  int
}

func NewRankingJob(client *rlock.Client, redisClient redis.Cmdable, nodeName string) *RankingJob {
	return &RankingJob{
		client:  client,
		key:     "rlock:cron_job:ranking",
		timeout: time.Minute * 1,
		job: func(ctx context.Context) error {
			fmt.Println(nodeName, time.Now().String())
			return nil
		},
		redisClient: redisClient,
		nodeName:    nodeName,
		threshhold:  1,
	}
}

func (r *RankingJob) Run() error {

	r.localLock.Lock()
	defer r.localLock.Unlock()

	res := r.redisClient.ZRank(context.Background(), "node_load", r.nodeName)
	rank := res.Val()
	// 获取节点负载 从redis zset 给出排序 负载不在前三名(threshhold)的结点 需要释放持有的分布式锁
	// 判断结点负载 节点负载高 释放锁
	if r.lock != nil && rank >= int64(r.threshhold) {
		fmt.Println(r.nodeName, " load too heavy, release redis lock")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		lock := r.lock
		r.lock = nil
		lock.Unlock(ctx)
		return nil
	}

	if r.lock == nil {
		// 尝试获取分布式锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      3,
		}, time.Second)
		// 尝试拿锁失败 直接返回了
		if err != nil {
			fmt.Println(err)
			return err
		}

		r.lock = lock

		// 续约 开新goroutine
		go func() {
			if r.lock != nil {
				err1 := r.lock.AutoRefresh(r.timeout/2, time.Second)
				if err1 != nil {
					r.localLock.Lock()
					r.lock = nil
					r.localLock.Unlock()
				}
			}

		}()

	}

	//申请到了分布式锁
	// 执行任务 在分布式锁过期时间内没有执行完任务需要进行分布式锁的续期
	// 分布式任务 RankingJob 最大执行时间 r.timeout 10min
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.job(ctx)

}

func (r RankingJob) Close() error {
	// 关机时释放锁
	r.localLock.Lock()
	lock := r.lock
	r.lock = nil
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)

}
