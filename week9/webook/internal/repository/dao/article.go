package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrPossibleIncorrectAuthor = errors.New("用户在尝试操作非本人数据")

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, aid int64, uid int64, status uint8) error
	GetById(ctx context.Context, id int64) (Article, error)
	GetPublishedById(ctx context.Context, id int64) (PublishedArticle, error)
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func (dao *GORMArticleDAO) GetPublishedById(ctx context.Context, id int64) (PublishedArticle, error) {
	var pubart PublishedArticle
	err := dao.db.Model(&PublishedArticle{}).WithContext(ctx).Where("id=?", id).First(&pubart).Error
	return pubart, err
}

func (dao *GORMArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).Where("id=?", id).First(&art).Error
	return art, err
}

func (dao *GORMArticleDAO) SyncStatus(ctx context.Context, aid int64, uid int64, status uint8) error {
	// transaction 把线上库制作库的status都设置为Private
	return dao.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id=? AND author_id=?", aid, uid).Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}
		res = tx.Model(&PublishedArticle{}).Where("id=? AND author_id=?", aid, uid).Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}
		return nil
	})
}

func (dao *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	now := time.Now().UnixMilli()
	err = dao.db.Transaction(func(tx *gorm.DB) error {
		// 在这个事务里 需要同时更新制作库和线上库
		txDAO := NewGORMArticleDAO(tx)

		// insert 语义
		if art.Id == 0 {
			id, err = txDAO.Insert(ctx, art)
		} else {
			err = txDAO.UpdateById(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		publishedArt := PublishedArticle(art)
		publishedArt.Utime = now
		publishedArt.Ctime = now
		err = tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":   art.Title,
				"content": art.Content,
				"status":  art.Status,
				"utime":   now,
			}),
		}).Create(&publishedArt).Error
		//if err != nil{
		//	return err
		//}

		return err
	})
	return id, err
}

func (dao *GORMArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()

	res := dao.db.Model(&Article{}).WithContext(ctx).Where("id = ? AND author_id=?", art.Id, art.AuthorId).Updates(
		map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("更新数据失败")
	}
	return nil
}

func (dao *GORMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Utime = now
	art.Ctime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

type PublishedArticle Article

type Article struct {
	//model
	Id int64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	// 标题的长度
	// 正常都不会超过这个长度
	Title   string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	Content string `gorm:"type=BLOB" bson:"content,omitempty"`
	// 作者
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8 `bson:"status,omitempty"`
	Ctime    int64 `bson:"ctime,omitempty"`
	Utime    int64 `bson:"utime,omitempty"`
}
