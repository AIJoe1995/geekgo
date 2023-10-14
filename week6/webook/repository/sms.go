package repository

import (
	"context"
	"github.com/ecodeclub/ekit/sqlx"
	"week6/webook/domain"
	"week6/webook/repository/dao"
)

// 短信的异步发送 存储数据库
var ErrWaitingSMSNotFound = dao.ErrWaitingSMSNotFound

type SMSRepository interface {
	Add(ctx context.Context, sms domain.SMS) error
	PreemptWaitingSMS(ctx context.Context) (domain.SMS, error)
	ReportScheduleResult(ctx context.Context, id int64, success bool) error
}

type smsRepository struct {
	dao dao.SMSDAO
}

func (s *smsRepository) PreemptWaitingSMS(ctx context.Context) (domain.SMS, error) {
	// 从数据库找到待发送的短信
	dao_sms, err := s.dao.GetWaitingSMS(ctx)
	return s.entityToDomain(dao_sms), err
}

func (s *smsRepository) Add(ctx context.Context, sms domain.SMS) error {
	return s.dao.Insert(ctx, s.domainToEntity(sms))
}

func (s *smsRepository) entityToDomain(sms dao.SMS) domain.SMS {
	return domain.SMS{
		Id:       sms.Id,
		TplId:    sms.Config.Val.TplId,
		Numbers:  sms.Config.Val.Numbers,
		Args:     sms.Config.Val.Args,
		RetryMax: sms.RetryMax,
	}
}

func (s *smsRepository) domainToEntity(sms domain.SMS) dao.SMS {
	return dao.SMS{
		Config: sqlx.JsonColumn[dao.SmsConfig]{
			Val: dao.SmsConfig{
				TplId:   sms.TplId,
				Args:    sms.Args,
				Numbers: sms.Numbers,
			},
			Valid: true,
		},
		RetryMax: sms.RetryMax,
	}
}

func (s *smsRepository) ReportScheduleResult(ctx context.Context, id int64, success bool) error {
	if success {
		return s.dao.MarkSuccess(ctx, id)
	}
	return s.dao.MarkFailed(ctx, id)
}
