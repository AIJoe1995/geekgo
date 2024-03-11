package repository

import (
	"context"
	"geekgo/week17/account/domain"
	"geekgo/week17/account/repository/dao"
	"gorm.io/gorm"
	"time"
)

type AccountRepository interface {
	AddCredit(ctx context.Context, c domain.Credit) error
}

type accountRepository struct {
	dao dao.AccountDAO
}

func (a *accountRepository) AddCredit(ctx context.Context, c domain.Credit) error {
	//
	//c.Items 更新账户动账记录 和 更新账户余额 需要在一个事务中完成
	err := a.dao.InsertCreditBizId(ctx, c.Biz, c.BizId)
	if err == gorm.ErrDuplicatedKey {
		return err
	} else {
		return err
	}

	activities := make([]dao.AccountActivity, 0, len(c.Items))
	now := time.Now().UnixMilli()
	for _, itm := range c.Items {
		activities = append(activities, dao.AccountActivity{
			Uid:         itm.Uid,
			Biz:         c.Biz,
			BizId:       c.BizId,
			Account:     itm.Account,
			AccountType: itm.AccountType.AsUint8(),
			Amount:      itm.Amt,
			Currency:    itm.Currency,
			Ctime:       now,
			Utime:       now,
		})
	}
	return a.dao.AddActivities(ctx, activities...)
}

func NewAccountRepository(dao dao.AccountDAO) AccountRepository {
	return &accountRepository{dao: dao}
}
