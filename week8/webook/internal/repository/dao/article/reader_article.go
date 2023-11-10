package article

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReaderDAO struct {
	db *gorm.DB
}

func (d *ReaderDAO) Save(ctx context.Context, art PublishedArticle) error {
	// 好像不能用Where upsert
	//return d.db.WithContext(ctx).Where("id=? AND author_id", art.Id, art.AuthorId).Clauses(
	//	clause.OnConflict{
	//		DoUpdates: clause.Assignments(map[string]any{
	//			"title": art.Title,
	//			"content": art.Content,
	//
	//		}),
	//	}).Create(&art).Error

	return d.db.Clauses(clause.OnConflict{
		// ID 冲突的时候。实际上，在 MYSQL 里面你写不写都可以
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
		}),
	}).Create(&art).Error
}

func NewReaderDAO(db *gorm.DB) *ReaderDAO {
	return &ReaderDAO{
		db: db,
	}
}

//type PublishedArticle Article
