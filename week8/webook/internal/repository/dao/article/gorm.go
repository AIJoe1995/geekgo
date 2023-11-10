package article

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrPossibleIncorrectAuthor = errors.New("用户在尝试操作非本人数据")

func NewGORMArticleDAO(db *gorm.DB) *GORMArticleDAO {
	return &GORMArticleDAO{db: db}
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func (d *GORMArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	res := d.db.Model(&Article{}).WithContext(ctx).Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"utime":   now,
		})
	//return res.Error
	err := res.Error
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return errors.New("更新数据失败")
	}
	return nil
}

func (d *GORMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := d.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (d *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	tx := d.db.WithContext(ctx).Begin()
	now := time.Now().UnixMilli()
	defer tx.Rollback()
	txDAO := NewArticleDAO(tx)
	var (
		id  = art.Id
		err error
	)
	if id == 0 {
		id, err = txDAO.Insert(ctx, art)
	} else {
		err = txDAO.UpdateById(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	publishArt := PublishedArticle(art)
	publishArt.Utime = now
	publishArt.Ctime = now
	err = tx.Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"utime":   now,
		}),
	}).Create(&publishArt).Error

	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, tx.Error
}

func (d *GORMArticleDAO) GetPublishedById(ctx context.Context, id int64) (PublishedArticle, error) {
	var pubArt PublishedArticle
	err := d.db.Where("id = ?", id).First(&pubArt).Error
	return pubArt, err
}

func (d *GORMArticleDAO) GetPublishedByIds(ctx context.Context, ids []int64) ([]PublishedArticle, error) {
	var pubArts []PublishedArticle
	err := d.db.Where("id in ?", ids).First(&pubArts).Error
	return pubArts, err
}

func (d *GORMArticleDAO) SyncStatus(ctx context.Context, author_id int64, article_id int64, status uint8) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id=? AND author_id = ?", article_id, author_id).
			Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}

		res = tx.Model(&PublishedArticle{}).
			Where("id=? AND author_id = ?", article_id, author_id).Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}
		return nil
	})

}

func NewArticleDAO(db *gorm.DB) *GORMArticleDAO {
	return &GORMArticleDAO{db: db}
}

type Article struct {
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

type PublishedArticle Article
