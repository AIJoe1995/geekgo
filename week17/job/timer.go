package job

import (
	"context"
	"geekgo/week17/payment/domain"
	"geekgo/week17/payment/repository"
	"time"
)

// 定时从PaymentEvent数据表里取出记录 进行重试 发送到kafka

type TickerProducer struct {
	repo          repository.PaymentRepository
	batchSize     int64
	sleepInterval time.Duration
}

func (t *TickerProducer) Produce() {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

		evts, err := t.repo.FindPaymentEvent(ctx, t.batchSize)
		if err != nil {
			continue
		}
		if len(evts) == 0 {
			time.Sleep(t.sleepInterval)
			continue
		}

		cancel()

		for _, evt := range evts {
			err2 := t.repo.CreatePaymentEvent(ctx, domain.PaymentEvent{
				BizTradeNO: evt.BizTradeNO,
				Status:     evt.Status,
			})
			if err2 != nil {
				err3 := t.repo.UpdatePaymentEventSent(ctx, evt.Id, 1)
				if err3 != nil {
					//记录日志
				}
			}
		}

	}

}
