package repository

import (
	"context"
	"geekgo/week19/follow/domain"
	"geekgo/week19/follow/repository/dao"
)

var _ FollowRepository = (*followRelationRepository)(nil)

type FollowRepository interface {
	// AddFollowRelation 创建关注关系
	AddFollowRelation(ctx context.Context, f domain.FollowRelation) error
	// InactiveFollowRelation 取消关注
	InactiveFollowRelation(ctx context.Context, follower int64, followee int64) error
	GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error)
	SetFollowerStatics(ctx context.Context, uid int64, delta int64) error
	SetFolloweeStatics(ctx context.Context, uid int64, delta int64) error
}

type followRelationRepository struct {
	dao dao.FollowRelationDAO
}

func NewFollowRelationRepository(dao dao.FollowRelationDAO) *followRelationRepository {
	return &followRelationRepository{dao: dao}
}

func (repo *followRelationRepository) SetFollowerStatics(ctx context.Context, uid int64, delta int64) error {
	return repo.dao.SetFollowerCnt(ctx, uid, delta)
}

func (repo *followRelationRepository) SetFolloweeStatics(ctx context.Context, uid int64, delta int64) error {
	return repo.dao.SetFolloweeCnt(ctx, uid, delta)
}

func (repo *followRelationRepository) AddFollowRelation(ctx context.Context, f domain.FollowRelation) error {
	err := repo.dao.CreateFollowRelation(ctx, repo.toEntity(f))
	if err != nil {
		return err
	}
	return err
}

func (repo *followRelationRepository) InactiveFollowRelation(ctx context.Context, follower int64, followee int64) error {
	err := repo.dao.UpdateStatus(ctx, followee, follower, dao.FollowRelationStatusInactive)
	if err != nil {
		return err
	}
	return err
}

func (repo *followRelationRepository) GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error) {

	res := domain.FollowStatics{Uid: uid}
	var err error
	res.Followers, err = repo.dao.CntFollower(ctx, uid)
	if err != nil {
		return res, err
	}
	res.Followees, err = repo.dao.CntFollowee(ctx, uid)
	return res, err
}

func (repo *followRelationRepository) toDomain(fr dao.FollowRelation) domain.FollowRelation {
	return domain.FollowRelation{
		Followee: fr.Followee,
		Follower: fr.Follower,
	}
}

func (repo *followRelationRepository) toEntity(c domain.FollowRelation) dao.FollowRelation {
	return dao.FollowRelation{
		Followee: c.Followee,
		Follower: c.Follower,
	}
}
