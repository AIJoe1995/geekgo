package repository

import (
	"context"
	"geekgo/week17/payment/domain"
	"geekgo/week17/payment/repository/dao"
)

var _ PaymentRepository = (*paymentRepository)(nil)

type PaymentRepository interface {
	AddPayment(ctx context.Context, pmt domain.Payment) error
	UpdatePayment(ctx context.Context, pmt domain.Payment) error
	CreatePaymentEvent(ctx context.Context, event domain.PaymentEvent) error
	FindPaymentEvent(ctx context.Context, limit int) ([]domain.PaymentEvent, error)
	UpdatePaymentEventSent(ctx context.Context, id int64, sentStatus int) error
}

type paymentRepository struct {
	dao dao.PaymentDAO
}

func (p *paymentRepository) UpdatePaymentEventSent(ctx context.Context, id int64, sentStatus int) error {
	return p.dao.UpdatePaymentEventSent(ctx, id, sentStatus)
}

func (p *paymentRepository) FindPaymentEvent(ctx context.Context, limit int) ([]domain.PaymentEvent, error) {
	res, err := p.dao.FindPaymentEvent(ctx, limit)
	if err != nil {
		return []domain.PaymentEvent{}, nil
	}
	res1 := make([]domain.PaymentEvent, 0, len(res))
	for _, r := range res {
		res1 = append(res1, domain.PaymentEvent{
			Id:         r.Id,
			BizTradeNO: r.BizTradeNO,
			Status:     r.Status,
			Sent:       r.Sent,
		})
	}
	return res1, nil
}

func (p *paymentRepository) CreatePaymentEvent(ctx context.Context, evt domain.PaymentEvent) error {
	return p.dao.InsertPaymentEvent(ctx, dao.PaymentEvent{
		BizTradeNO: evt.BizTradeNO,
		Status:     evt.Status,
	})
}

func (p *paymentRepository) toEntity(pmt domain.Payment) dao.Payment {
	return dao.Payment{
		Amt:         pmt.Amt.Total,
		Currency:    pmt.Amt.Currency,
		BizTradeNO:  pmt.BizTradeNO,
		Description: pmt.Description,
		Status:      domain.PaymentStatusInit,
	}
}

func (p *paymentRepository) toDomain(pmt dao.Payment) domain.Payment {
	return domain.Payment{
		Amt: domain.Amount{
			Currency: pmt.Currency,
			Total:    pmt.Amt,
		},
		BizTradeNO:  pmt.BizTradeNO,
		Description: pmt.Description,
		Status:      domain.PaymentStatus(pmt.Status),
		TxnID:       pmt.TxnID.String,
	}
}

func (p *paymentRepository) AddPayment(ctx context.Context, pmt domain.Payment) error {
	return p.dao.Insert(ctx, p.toEntity(pmt))
}

func (p *paymentRepository) UpdatePayment(ctx context.Context, pmt domain.Payment) error {
	return p.dao.UpdateTxnIDAndStatus(ctx, pmt.BizTradeNO, pmt.TxnID, pmt.Status)
}
