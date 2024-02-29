package job

import (
	"context"
	"encoding/json"
	"fmt"
	"geekgo/week11/domain"
	"geekgo/week11/service"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
	"net/http"
	"time"
)

type Scheduler struct {
	execs   map[string]Executor    // 执行任务
	svc     service.CronJobService // 负责抢占任务
	limiter *semaphore.Weighted    // 限制同一节点抢占任务个数
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.execs[exec.Name()] = exec
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			// 退出了Schedule 什么时候再重新进行Schedule??
			return ctx.Err()
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			// 出现错误退出了schedule循环 是不是应该处理semaphore
			return err
		}
		// 数据库查询的时间 dbCtx
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			// 抢占任务失败 继续下一轮抢占
			continue
		}
		exec, ok := s.execs[j.Executor]
		if !ok {
			// 没有找到 执行任务的executor 线下 return err 线上continue
			continue
		}
		// 开启单独goroutine执行任务
		go func() {
			defer func() {
				s.limiter.Release(1)
				err1 := j.CancelFunc()
				if err1 != nil {
					// 记录日志 释放任务失败
				}
			}()
			err1 := exec.Exec(ctx, j)
			if err1 != nil {
				// 执行任务失败 记录日志 可以重试
			}

			// 任务执行完毕 需要设置下一次任务运行时间
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			s.svc.ResetNextTime(ctx, j)
			if err1 != nil {
				// 设置下一次执行时间失败 记录日志
			}
		}()

	}
}

// Executor 接口 提供不同实现 注册到Schedular的execs map
type Executor interface {
	Name() string
	Exec(ctx context.Context, j domain.CronJob) error
}

type HttpExecutor struct {
}

func (h *HttpExecutor) Name() string {
	return "http"
}

func (h *HttpExecutor) Exec(ctx context.Context, j domain.CronJob) error {
	type Config struct {
		Endpoint string
		Method   string
	}
	var cfg Config
	err := json.Unmarshal([]byte(j.Cfg), &cfg)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(cfg.Method, cfg.Endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("%s 任务执行失败", j.Name)
	}
	return nil
}

type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.CronJob) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{
		funcs: make(map[string]func(ctx context.Context, j domain.CronJob) error),
	}
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, j domain.CronJob) error) {
	l.funcs[name] = fn
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.CronJob) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未知任务，你是否注册了？ %s", j.Name)
	}
	return fn(ctx, j)
}

func NewScheduler(svc service.CronJobService) *Scheduler {
	return &Scheduler{svc: svc,
		limiter: semaphore.NewWeighted(200),
		execs:   make(map[string]Executor)}
}
