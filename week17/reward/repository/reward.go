package repository

import (
	"context"
	"geekgo/week17/reward/domain"
	"geekgo/week17/reward/repository/cache"
	"geekgo/week17/reward/repository/dao"
)

var _ RewardRepository = (*rewardRepository)(nil)

type RewardRepository interface {
	CreateReward(ctx context.Context, r domain.Reward) (int64, error)
	GetReward(ctx context.Context, rid int64) (domain.Reward, error)
	UpdateStatus(ctx context.Context, rid int64, status domain.RewardStatus) error
}

type rewardRepository struct {
	dao   dao.RewardDAO
	cache cache.RewardCache
}

func (repo *rewardRepository) GetReward(ctx context.Context, rid int64) (domain.Reward, error) {
	r, err := repo.dao.GetReward(ctx, rid)
	if err != nil {
		return domain.Reward{}, err
	}
	return repo.toDomain(r), nil
}

func (repo *rewardRepository) UpdateStatus(ctx context.Context, rid int64, status domain.RewardStatus) error {

	return repo.dao.UpdateStatus(ctx, rid, status.AsUint8())
}

func (repo *rewardRepository) CreateReward(ctx context.Context, r domain.Reward) (int64, error) {
	return repo.dao.Insert(ctx, repo.toEntity(r))
}

func (repo *rewardRepository) toEntity(r domain.Reward) dao.Reward {
	return dao.Reward{
		Status:    r.Status.AsUint8(),
		Biz:       r.Target.Biz,
		BizName:   r.Target.BizName,
		BizId:     r.Target.BizId,
		TargetUid: r.Target.Uid,
		Uid:       r.Uid,
		Amount:    r.Amt,
	}
}

func (repo *rewardRepository) toDomain(r dao.Reward) domain.Reward {
	return domain.Reward{
		Id:  r.Id,
		Uid: r.Uid,
		Target: domain.Target{
			Biz:     r.Biz,
			BizId:   r.BizId,
			BizName: r.BizName,
			Uid:     r.Uid,
		},
		Amt:    r.Amount,
		Status: domain.RewardStatus(r.Status),
	}
}
