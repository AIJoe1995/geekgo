package failoverWithRatelimit

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
	"week6/webook/domain"
	"week6/webook/pkg/ratelimit"
	"week6/webook/repository"
	"week6/webook/service/sms"
)

type Service struct {
	svcs []sms.Service
	idx  int32

	avgReqTime           int64 // 毫秒数
	avgReqTimeThreshhold int64 // 毫秒数
	reqNum               int64

	limiter ratelimit.Limiter
	repo    repository.SMSRepository
}

func (s *Service) Send(ctx context.Context, tpl string, args []string, phone []string) error {
	newidx := s.idx
	length := int32(len(s.svcs))
	// failover后使用下一个idx的服务商发送
	for idx := s.idx; idx < s.idx+length; idx++ {
		newidx = idx % length
		if s.avgReqTime >= s.avgReqTimeThreshhold {
			// 切换服务商
			// 结构体内使用了atomic.Int32类型 或int32类型 Send里对这两种类型的操作
			idx := s.idx
			newidx = (idx + 1) % length

			if atomic.CompareAndSwapInt32(&s.idx, s.idx, newidx) {
				atomic.StoreInt64(&s.avgReqTime, 0)
				atomic.StoreInt64(&s.reqNum, 0)
			}
		}

		limited, err := s.limiter.Limit(ctx, tpl)
		if limited {
			// 被限流要转异步发送短信 需要在service里面继承repo层 把request保存到短信数据库里 重新取出发送

			_ = s.repo.Add(ctx, domain.SMS{
				TplId:   tpl,
				Args:    args,
				Numbers: phone,
				// 设置可以重试三次
				RetryMax: 3,
			})

			return errors.New("限流，短信发送失败")
		}
		if err != nil {
			return err
		}

		start := time.Now().UnixMilli()
		err = s.svcs[newidx].Send(ctx, tpl, args, phone)
		if err != nil {
			return err
		}
		end := time.Now().UnixMilli()
		reqTotalTime := (atomic.LoadInt64(&s.avgReqTime) * atomic.LoadInt64(&s.reqNum)) + (end - start)
		atomic.AddInt64(&s.reqNum, 1)
		reqAvgTime := reqTotalTime / atomic.LoadInt64(&s.reqNum)
		atomic.CompareAndSwapInt64(&s.avgReqTime, atomic.LoadInt64(&s.avgReqTime), reqAvgTime)
		switch err {
		case nil:
			return nil
		case context.DeadlineExceeded, context.Canceled:
			return err
			//default:
			//	// 输出日志
			//	return err
		}
	}
	// 如果所有服务商都崩溃了 需要转异步 这里向数据库插入短信如果失败了 要怎么处理？？
	_ = s.repo.Add(ctx, domain.SMS{
		TplId:   tpl,
		Args:    args,
		Numbers: phone,
		// 设置可以重试三次
		RetryMax: 3,
	})

	return errors.New("全部服务商都失败了")
}
