package ratelimit

import (
	"context"
	"errors"
	"week6/webook/pkg/ratelimit"
	"week6/webook/service/sms"
)

// 实现带有限流的短信服务 通过装饰器修饰sms.Service接口
// 逻辑 如果当前限流 那么不发送短信 返回限流错误 不限流正常发送短信

type Service struct {
	smsSvc  sms.Service
	limiter ratelimit.Limiter
}

func (s *Service) Send(ctx context.Context, tpl string, args []string, phone []string) error {
	// 是否限流通过中间件来判断， 逻辑是 一段时间窗口的请求数超过某阈值 就执行限流 调用接口需要先新增加一个请求 如果一段时间内已经很多请求累计，那这个请求就被丢弃
	limited, err := s.limiter.Limit(ctx, tpl)
	if err != nil {
		return err
	}
	if limited {
		return errors.New("限流，发送短信失败")
	}
	return s.smsSvc.Send(ctx, tpl, args, phone)
}
