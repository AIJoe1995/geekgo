package dao

import (
	"geekgo/week8/webook/internal/repository/dao/article"
	"gorm.io/gorm"
)

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(&User{},
		&article.Article{},
		&article.PublishedArticle{},
		&Interactive{},
		&UserLikeBiz{},
		&Collection{},
		&UserCollectionBiz{},
	)
}
