package service

import (
	"context"
	"geekgo/week11/domain"
	"geekgo/week11/repository"
	"time"
)

type CronJobService interface {
	AddJob(ctx context.Context, job domain.CronJob) error
	Preempt(ctx context.Context) (domain.CronJob, error)
	ResetNextTime(ctx context.Context, job domain.CronJob) error
}

type cronJobService struct {
	// 调用repository层的方法 操作数据库 向数据库中查询插入更新删除
	repo repository.CronJobRepository

	// 抢占到任务的结点 需要定时向数据库刷新 作为健康证明
	refreshInterval time.Duration
}

func (c *cronJobService) AddJob(ctx context.Context, job domain.CronJob) error {
	job.NextTime = job.Next(time.Now())
	return c.repo.AddJob(ctx, job)
}

func (c *cronJobService) healthCheck(id int64, ch chan struct{}) {
	ticker := time.NewTicker(c.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.refresh(id)
		case <-ch:
			return
		}
	}
}

func (c *cronJobService) refresh(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.repo.UpdateUtime(ctx, id)
	if err != nil {
		// 记录日志
	}
}

func (c *cronJobService) Preempt(ctx context.Context) (domain.CronJob, error) {
	// 从数据库抢占到一个任务 在任务调度模块里 可以开启多个goroutine同时进行Preempt抢占操作
	job, err := c.repo.Preempt(ctx)
	if err != nil {
		return domain.CronJob{}, err
	}

	ch := make(chan struct{})

	// 抢占到任务 需要 不断更新Utime 证明结点活跃  当取消任务的函数被调用时 需要通知更新Utime的函数 停止更新Utime
	go c.healthCheck(job.Id, ch)

	job.CancelFunc = func() error {
		return c.createCancelFunc(job.Id, ch)
	}
	return job, nil

}

func (c *cronJobService) ResetNextTime(ctx context.Context, job domain.CronJob) error {
	t := job.Next(time.Now())
	if !t.IsZero() {
		return c.repo.UpdateNextTime(ctx, job.Id, t)
	} else {
		// 应该标记为 任务已经完成
		return c.repo.EndJob(ctx, job.Id)
	}
	return nil
}

func (c *cronJobService) createCancelFunc(id int64, ch chan struct{}) error {
	// 任务取消 需要通知healthCheck 停止 更新Utime 停止refresh
	// 任务取消 需要将mysql表中记录的状态从被强占修改为可以被抢占
	close(ch)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.repo.Release(ctx, id)
	if err != nil {
		// 释放任务失败
		return err
	}
	return nil

}

func NewCronJobService(repo repository.CronJobRepository) CronJobService {
	return &cronJobService{
		repo:            repo,
		refreshInterval: time.Second * 10,
	}
}
