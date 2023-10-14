package async

import (
	"context"
	"fmt"
	"time"
	"week6/webook/repository"
	"week6/webook/service/sms"
)

// 调用failoverWithRatelimit的发送短信被限流或服务商都崩溃时，短信被存储到数据库了
// 这里开启一个异步服务从数据库中找到 待发送的短信 来发送

type AsyncSMSService interface {
	AsyncSendSMS()
}

type asyncSMSService struct {
	svc  sms.Service
	repo repository.SMSRepository
}

func (a *asyncSMSService) AsyncSendSMS() {
	// 从数据库取出短信
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 抢占一个异步发送的消息，确保在非常多个实例
	// 比如 k8s 部署了三个 pod，一个请求，只有一个实例能拿到
	as, err := a.repo.PreemptWaitingSMS(ctx)
	cancel()
	// 如果没有err就发送，如果库里没有需要发送的短信了 就sleep 如果是其他err 就返回err
	switch err {
	case nil:
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err := a.svc.Send(ctx, as.TplId, as.Args, as.Numbers)
		err = a.repo.ReportScheduleResult(ctx, as.Id, err == nil)
	case repository.ErrWaitingSMSNotFound:
		time.Sleep(time.Second)
	default:
		fmt.Println(err)
	}

}

func (a *asyncSMSService) StartAsyncSend() {
	for {
		a.AsyncSendSMS()
	}
}

// 在启动服务时启动go routine
func NewAsyncSMSService(svc sms.Service, repo repository.SMSRepository) AsyncSMSService {
	async := &asyncSMSService{
		svc:  svc,
		repo: repo,
	}
	go func() {
		async.StartAsyncSend()
	}()
	return async
}
