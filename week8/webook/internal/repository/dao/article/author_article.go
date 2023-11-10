package article

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type AuthorDAO struct {
	db *gorm.DB
}

func (d *AuthorDAO) Create(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := d.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (d *AuthorDAO) Update(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	return d.db.WithContext(ctx).Where("id=? AND author_id=?", art.Id, art.AuthorId).Updates(map[string]any{
		"title":   art.Title,
		"content": art.Content,
		"utime":   now,
	}).Error
}

func NewAuthorDAO(db *gorm.DB) *AuthorDAO {
	return &AuthorDAO{
		db: db,
	}
}
