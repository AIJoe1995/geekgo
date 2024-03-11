package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

var _ RewardDAO = (*GORMRewardDAO)(nil)

type (
	RewardDAO interface {
		Insert(ctx context.Context, r Reward) (int64, error)
		GetReward(ctx context.Context, rid int64) (Reward, error)
		UpdateStatus(ctx context.Context, rid int64, status uint8) error
	}
)

type GORMRewardDAO struct {
	db *gorm.DB
}

func (g *GORMRewardDAO) GetReward(ctx context.Context, rid int64) (Reward, error) {
	var r Reward
	err := g.db.WithContext(ctx).Where("id=?", rid).Find(&r).Error
	return r, err
}

func (g *GORMRewardDAO) UpdateStatus(ctx context.Context, rid int64, status uint8) error {
	return g.db.WithContext(ctx).
		Where("id = ?", rid).
		Updates(map[string]any{
			"status": status,
			"utime":  time.Now().UnixMilli(),
		}).Error
}

func (g *GORMRewardDAO) Insert(ctx context.Context, r Reward) (int64, error) {
	now := time.Now().UnixMilli()
	r.Ctime = now
	r.Utime = now
	err := g.db.WithContext(ctx).Create(&r).Error
	return r.Id, err
}
