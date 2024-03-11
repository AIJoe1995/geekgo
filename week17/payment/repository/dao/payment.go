package dao

import (
	"context"
	"geekgo/week17/payment/domain"
	"gorm.io/gorm"
	"time"
)

var _ PaymentDAO = (*GORMPaymentDAO)(nil)

type PaymentDAO interface {
	Insert(ctx context.Context, pmt Payment) error
	UpdateTxnIDAndStatus(ctx context.Context, bizTradeNo string, txnId string, status domain.PaymentStatus) error
	InsertPaymentEvent(ctx context.Context, event PaymentEvent) error
	FindPaymentEvent(ctx context.Context, limit int) ([]PaymentEvent, error)
	UpdatePaymentEventSent(ctx context.Context, id int64, sent int) error
}

type GORMPaymentDAO struct {
	db *gorm.DB
}

func (g *GORMPaymentDAO) UpdatePaymentEventSent(ctx context.Context, id int64, sent int) error {
	return g.db.WithContext(ctx).Model(&PaymentEvent{}).Where("id=?", id).
		Updates(map[string]interface{}{
			"sent":  1,
			"utime": time.Now().UnixMilli(),
		}).Error
}

func (g *GORMPaymentDAO) FindPaymentEvent(ctx context.Context, limit int) ([]PaymentEvent, error) {
	res := make([]PaymentEvent, 0, limit)
	err := g.db.WithContext(ctx).Model(&PaymentEvent{}).Where("sent != 1 ").Order("utime").Limit(limit).Find(&res).Error
	return res, err
}

func (g *GORMPaymentDAO) InsertPaymentEvent(ctx context.Context, evt PaymentEvent) error {
	now := time.Now().UnixMilli()
	evt.Ctime = now
	evt.Utime = now
	return g.db.WithContext(ctx).Create(&evt).Error
}

func (g *GORMPaymentDAO) Insert(ctx context.Context, pmt Payment) error {
	now := time.Now().UnixMilli()
	pmt.Utime = now
	pmt.Ctime = now
	return g.db.WithContext(ctx).Create(&pmt).Error
}

func (g *GORMPaymentDAO) UpdateTxnIDAndStatus(ctx context.Context, bizTradeNo string, txnId string, status domain.PaymentStatus) error {
	return g.db.WithContext(ctx).Model(&Payment{}).Where("biz_trade_no=?", bizTradeNo).
		Updates(map[string]any{
			"txn_id": txnId,
			"status": status.AsUint8(),
			"utime":  time.Now().UnixMilli(),
		}).Error
}
