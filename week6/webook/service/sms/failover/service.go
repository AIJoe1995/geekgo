package failover

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
	"week6/webook/service/sms"
)

//故障转移
// 当发现一个服务商发生故障时，使用其他服务商提供服务
// 故障的判定方法可以是 服务商的过去一段时间内的请求失败率， 请求失败率超过阈值 则更换服务商
// 其他故障判定方法 如 平均响应时间超过阈值 一段时间内99%的用户响应时间应该在阈值内等等
//

type Service struct {
	svcs []sms.Service
	idx  int32

	avgReqTime           int64 // 毫秒数
	avgReqTimeThreshhold int64 // 毫秒数
	reqNum               int64
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
			}

		}
		start := time.Now().UnixMilli()
		err := s.svcs[newidx].Send(ctx, tpl, args, phone)
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
		default:
			// 输出日志
			return err
		}
	}
	return errors.New("全部服务商都失败了")
}
