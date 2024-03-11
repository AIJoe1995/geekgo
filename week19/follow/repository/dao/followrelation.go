package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var _ FollowRelationDAO = (*GORMFollowRelationDAO)(nil)

type FollowRelation struct {
	ID int64 `gorm:"primaryKey,autoIncrement,column:id"`

	Follower int64 `gorm:"type:int(11);not null;uniqueIndex:follower_followee"`
	Followee int64 `gorm:"type:int(11);not null;uniqueIndex:follower_followee"`

	Status uint8

	// 这里你可以根据自己的业务来增加字段，比如说
	// 关系类型，可以搞些什么普通关注，特殊关注
	// Type int64 `gorm:"column:type;type:int(11);comment:关注类型 0-普通关注"`
	// 备注
	// Note string `gorm:"column:remark;type:varchar(255);"`
	// 创建时间
	Ctime int64
	Utime int64
}

type FollowStatics struct {
	ID  int64 `gorm:"primaryKey,autoIncrement,column:id"`
	Uid int64 `gorm:"unique"`
	// 有多少粉丝
	Followers int64
	// 关注了多少人
	Followees int64

	Utime int64
	Ctime int64
}

func (FollowRelation) TableName() string {
	return "follow_relations"
}

func (FollowStatics) TableName() string {
	return "follow_statics"
}

const (
	FollowRelationStatusUnknown uint8 = iota
	FollowRelationStatusActive
	FollowRelationStatusInactive
)

type FollowRelationDAO interface {
	// CreateFollowRelation 创建联系人
	CreateFollowRelation(ctx context.Context, c FollowRelation) error
	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, followee int64, follower int64, status uint8) error
	// CntFollower 统计计算关注自己的人有多少
	CntFollower(ctx context.Context, uid int64) (int64, error)
	// CntFollowee 统计自己关注了多少人
	CntFollowee(ctx context.Context, uid int64) (int64, error)

	SetFollowerCnt(ctx context.Context, uid int64, delta int64) error
	SetFolloweeCnt(ctx context.Context, uid int64, delta int64) error
}

type GORMFollowRelationDAO struct {
	db *gorm.DB
}

func NewGORMFollowRelationDAO(db *gorm.DB) *GORMFollowRelationDAO {
	return &GORMFollowRelationDAO{db: db}
}

func (dao *GORMFollowRelationDAO) SetFollowerCnt(ctx context.Context, uid int64, delta int64) error {
	return dao.db.WithContext(ctx).Model(&FollowStatics{}).Where("uid=?", uid).Updates(map[string]interface{}{
		"followers": gorm.Expr("followers + %d", delta),
		"utime":     time.Now().UnixMilli(),
	}).Error
}

func (dao *GORMFollowRelationDAO) SetFolloweeCnt(ctx context.Context, uid int64, delta int64) error {
	return dao.db.WithContext(ctx).Model(&FollowStatics{}).Where("uid=?", uid).Updates(map[string]interface{}{
		"followees": gorm.Expr("followees + %d", delta),
		"utime":     time.Now().UnixMilli(),
	}).Error
}

func (dao *GORMFollowRelationDAO) CreateFollowRelation(ctx context.Context, f FollowRelation) error {
	now := time.Now().UnixMilli()
	f.Utime = now
	f.Ctime = now
	f.Status = FollowRelationStatusActive
	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]any{
			"utime":  now,
			"status": FollowRelationStatusActive,
		}),
	}).Create(&f).Error
}

func (dao *GORMFollowRelationDAO) UpdateStatus(ctx context.Context, followee int64, follower int64, status uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).
		Where("follower = ? AND followee = ?", follower, followee).
		Updates(map[string]any{
			"status": status,
			"utime":  now,
		}).Error
}

func (dao *GORMFollowRelationDAO) CntFollower(ctx context.Context, uid int64) (int64, error) {
	var res int64
	err := dao.db.WithContext(ctx).Model(&FollowRelation{}).
		Select("count(follower)").
		Where("followee = ? AND status = ?",
			uid, FollowRelationStatusActive).Count(&res).Error
	return res, err
}

func (dao *GORMFollowRelationDAO) CntFollowee(ctx context.Context, uid int64) (int64, error) {
	var res int64
	err := dao.db.WithContext(ctx).Model(&FollowRelation{}).
		Select("count(followee)").
		Where("follower = ? AND status = ?",
			uid, FollowRelationStatusActive).Count(&res).Error
	return res, err
}
