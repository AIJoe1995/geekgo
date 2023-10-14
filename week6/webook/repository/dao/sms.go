package dao

import (
	"context"
	"github.com/ecodeclub/ekit/sqlx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrWaitingSMSNotFound = gorm.ErrRecordNotFound

const (
	// 因为本身状态没有暴露出去，所以不需要在 domain 里面定义
	asyncStatusWaiting = iota
	// 失败了，并且超过了重试次数
	asyncStatusFailed
	asyncStatusSuccess
)

type SMSDAO interface {
	Insert(ctx context.Context, sms SMS) error
	MarkSuccess(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
	GetWaitingSMS(ctx context.Context) (SMS, error)
}

type GORMAsyncSmsDAO struct {
	db *gorm.DB
}

func (g *GORMAsyncSmsDAO) Insert(ctx context.Context, sms SMS) error {
	return g.db.Create(&sms).Error
}

func (g *GORMAsyncSmsDAO) GetWaitingSMS(ctx context.Context) (SMS, error) {
	var s SMS
	err := g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 为了避开一些偶发性的失败，我们只找 1 分钟前的异步短信发送
		now := time.Now().UnixMilli()
		endTime := now - time.Minute.Milliseconds()
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("utime < ? and status = ?",
				endTime, asyncStatusWaiting).First(&s).Error
		// SELECT xx FROM xxx WHERE xx FOR UPDATE，锁住了
		if err != nil {
			return err
		}

		// 只要更新了更新时间，根据我们前面的规则，就不可能被别的节点抢占了
		err = tx.Model(&SMS{}).
			Where("id = ?", s.Id).
			Updates(map[string]any{
				"retry_cnt": gorm.Expr("retry_cnt + 1"),
				// 更新成了当前时间戳，确保我在发送过程中，没人会再次抢到它
				// 也相当于，重试间隔一分钟
				"utime": now,
			}).Error
		return err
	})
	return s, err
}

func (g *GORMAsyncSmsDAO) MarkSuccess(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Model(&SMS{}).
		Where("id =?", id).
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusSuccess,
		}).Error
}

func (g *GORMAsyncSmsDAO) MarkFailed(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Model(&SMS{}).
		// 只有到达了重试次数才会更新
		Where("id =? and `retry_cnt`>=`retry_max`", id).
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusFailed,
		}).Error
}

type SMS struct {
	Id     int64
	Config sqlx.JsonColumn[SmsConfig]
	// 设置可以重试三次
	RetryCnt int
	RetryMax int
	// 重试的最大次数
	Status uint8
	Ctime  int64
	Utime  int64 `gorm:"index"`
}

type SmsConfig struct {
	TplId   string
	Args    []string
	Numbers []string
}
