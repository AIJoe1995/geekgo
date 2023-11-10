package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, id int64) error
	InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error
	DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error
	InsertCollectInfo(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (Interactive, error)
	GetLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserLikeBiz, error)
	GetCollectInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserCollectionBiz, error)
	TopNLike(ctx context.Context, biz string, num int64) ([]Interactive, error)
}

type GORMInteractiveDAO struct {
	db *gorm.DB
}

func (dao *GORMInteractiveDAO) TopNLike(ctx context.Context, biz string, num int64) ([]Interactive, error) {
	// biz=biz order by like_cnt limit topn
	intrs := []Interactive{}
	err := dao.db.Limit(int(num)).Where("biz=?", biz).Order(clause.OrderByColumn{Column: clause.Column{Name: "like_cnt"}, Desc: true}).Find(&intrs).Error
	if err != nil {
		return []Interactive{}, nil
	}
	return intrs, err
}

func (dao *GORMInteractiveDAO) GetLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := dao.db.WithContext(ctx).
		Where("biz=? AND biz_id = ? AND uid = ? AND status = ?",
			biz, bizId, uid, 1).First(&res).Error
	return res, err
}

func (dao *GORMInteractiveDAO) GetCollectInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := dao.db.WithContext(ctx).
		Where("biz=? AND biz_id = ? AND uid = ?", biz, bizId, uid).First(&res).Error
	return res, err
}

func (dao *GORMInteractiveDAO) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	var intr Interactive
	err := dao.db.WithContext(ctx).Where("biz = ? AND biz_id = ?", biz, bizId).First(&intr).Error
	return intr, err
}

func (dao *GORMInteractiveDAO) InsertCollectInfo(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	//
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 要是还没有该收藏夹cid 怎么办
			err := dao.db.WithContext(ctx).Create(&UserCollectionBiz{
				Cid:   cid,
				Biz:   biz,
				BizId: bizId,
				Uid:   uid,
				Ctime: now,
				Utime: now,
			}).Error
			if err != nil {
				return err
			}
			return tx.Clauses(clause.OnConflict{
				DoUpdates: clause.Assignments(map[string]any{
					"collect_cnt": gorm.Expr("`collect_cnt` + 1"),
					"utime":       now,
				}),
			}).Create(&Interactive{
				CollectCnt: 1,
				Ctime:      now,
				Utime:      now,
				Biz:        biz,
				BizId:      bizId,
			}).Error

		})
}

func (dao *GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	// // 同时记录点赞，以及更新点赞计数
	// delete和insert的区别 insertlikeinfo需要使用upsert语义，没找到新建，而delete没找到不需要新建
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 两个操作
		// 一个是软删除点赞记录
		// 一个是减点赞数量
		err := tx.Model(&UserLikeBiz{}).
			Where("biz=? AND biz_id = ? AND uid = ?", biz, bizId, uid).
			Updates(map[string]any{
				"utime":  now,
				"status": 0,
			}).Error
		if err != nil {
			return err
		}
		return tx.Model(&Interactive{}).
			// 这边命中了索引，然后没找到，所以不会加锁
			Where("biz=? AND biz_id = ?", biz, bizId).
			Updates(map[string]any{
				"utime":    now,
				"like_cnt": gorm.Expr("like_cnt-1"),
			}).Error
	})

	//return dao.db.WithContext(ctx).Transaction(
	//	// transaction 传入一个函数作为参数，这个函数的参数不用管，Transaction内部会调用fn时给fn传参
	//	func(tx *gorm.DB) error {
	//		err := tx.Clauses(clause.OnConflict{
	//			DoUpdates: clause.Assignments(map[string]any{
	//				"utime":  now,
	//				"status": 0,
	//			}),
	//		}).Create(&UserLikeBiz{
	//			Biz:    biz,
	//			BizId:  bizId,
	//			Uid:    uid,
	//			Status: 0,
	//			Ctime:  now,
	//			Utime:  now,
	//		}).Error
	//		if err != nil {
	//			return err
	//		}
	//		return tx.Clauses(clause.OnConflict{
	//			DoUpdates: clause.Assignments(map[string]any{
	//				"like_cnt": gorm.Expr("like_cnt - 1"), // like_cnt是null会怎么样 是不是应该在like_cnt上设置default 0
	//				"utime":    time.Now().UnixMilli(),
	//			}),
	//		}).Create(&Interactive{
	//			Biz:     biz,
	//			BizId:   bizId,
	//			LikeCnt: 0,
	//			Ctime:   now,
	//			Utime:   now,
	//		}).Error
	//	})
}

func (dao *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	// // 同时记录点赞，以及更新点赞计数
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(
		// transaction 传入一个函数作为参数，这个函数的参数不用管，Transaction内部会调用fn时给fn传参
		func(tx *gorm.DB) error {
			err := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.Assignments(map[string]any{
					"utime":  now,
					"status": 1,
				}),
			}).Create(&UserLikeBiz{
				Biz:    biz,
				BizId:  bizId,
				Uid:    uid,
				Status: 1,
				Ctime:  now,
				Utime:  now,
			}).Error
			if err != nil {
				return err
			}
			return tx.Clauses(clause.OnConflict{
				DoUpdates: clause.Assignments(map[string]any{
					"like_cnt": gorm.Expr("like_cnt + 1"), // like_cnt是null会怎么样 是不是应该在like_cnt上设置default 0
					"utime":    time.Now().UnixMilli(),
				}),
			}).Create(&Interactive{
				Biz:     biz,
				BizId:   bizId,
				LikeCnt: 1,
				Ctime:   now,
				Utime:   now,
			}).Error
		})
}

func NewGORMInteractiveDAO(db *gorm.DB) *GORMInteractiveDAO {
	return &GORMInteractiveDAO{db: db}
}

func (dao *GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			DoUpdates: clause.Assignments(
				map[string]any{
					"read_cnt": gorm.Expr("read_cnt + 1"),
					"utime":    time.Now().UnixMilli(),
				}),
		}).Create(&Interactive{
		Biz:     biz,
		BizId:   bizId,
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
	}).Error
}

type Interactive struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 业务标识符
	// 同一个资源，在这里应该只有一行
	// 也就是说我要在 bizId 和 biz 上创建联合唯一索引
	// 1. bizId, biz。优先选择这个，因为 bizId 的区分度更高
	// 2. biz, bizId。如果有 WHERE biz = xx 这种查询条件（不带 bizId）的，就只能这种
	//
	// 联合索引的列的顺序：查询条件，区分度
	// 这个名字无所谓
	BizId int64 `gorm:"uniqueIndex:biz_id_type"`
	// 我这里biz 用的是 string，有些公司枚举使用的是 int 类型
	// 0-article
	// 1- xxx
	// 默认是 BLOB/TEXT 类型
	Biz string `gorm:"uniqueIndex:biz_id_type;type:varchar(128)"`
	// 这个是阅读计数
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Ctime      int64
	Utime      int64
}

// UserLikeBiz 命名无能，用户点赞的某个东西
type UserLikeBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`

	// 我在前端展示的时候，
	// WHERE uid = ? AND biz_id = ? AND biz = ?
	// 来判定你有没有点赞
	// 这里，联合顺序应该是什么？

	// 要分场景
	// 1. 如果你的场景是，用户要看看自己点赞过那些，那么 Uid 在前
	// WHERE uid =?
	// 2. 如果你的场景是，我的点赞数量，需要通过这里来比较/纠正
	// biz_id 和 biz 在前
	// select count(*) where biz = ? and biz_id = ?
	Biz   string `gorm:"uniqueIndex:uid_biz_id_type;type:varchar(128)"`
	BizId int64  `gorm:"uniqueIndex:uid_biz_id_type"`

	// 谁的操作
	Uid int64 `gorm:"uniqueIndex:uid_biz_id_type"`

	Ctime int64
	Utime int64
	// 如果这样设计，那么，取消点赞的时候，怎么办？
	// 我删了这个数据
	// 你就软删除
	// 这个状态是存储状态，纯纯用于软删除的，业务层面上没有感知
	// 0-代表删除，1 代表有效
	Status uint8

	// 有效/无效
	//Type string
}

// Collection 收藏夹
type Collection struct {
	Id   int64  `gorm:"primaryKey,autoIncrement"`
	Name string `gorm:"type=varchar(1024)"`
	Uid  int64  `gorm:""`

	Ctime int64
	Utime int64
}

// UserCollectionBiz 收藏的东西
type UserCollectionBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 收藏夹 ID
	// 作为关联关系中的外键，我们这里需要索引
	Cid   int64  `gorm:"index"`
	BizId int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_uid"`
	// 这算是一个冗余，因为正常来说，
	// 只需要在 Collection 中维持住 Uid 就可以
	Uid   int64 `gorm:"uniqueIndex:biz_type_id_uid"`
	Ctime int64
	Utime int64
}

// 假如说我有一个需求，需要查询到收藏夹的信息，和收藏夹里面的资源
// SELECT c.id as cid , c.name as cname, uc.biz_id as biz_id, uc.biz as biz
// FROM `collection` as c JOIN `user_collection_biz` as uc
// ON c.id = uc.cid
// WHERE c.id IN (1,2,3)

type CollectionItem struct {
	Cid   int64
	Cname string
	BizId int64
	Biz   string
}
