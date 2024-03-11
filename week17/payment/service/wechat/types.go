package wechat

import (
	"context"
	"geekgo/week17/payment/domain"
)

type PaymentService interface {
	Prepay(ctx context.Context, pmt domain.Payment) (string, error)
}
