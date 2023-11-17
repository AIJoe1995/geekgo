package repository

import (
	"context"
	"geekgo/week9/webook/internal/domain"
	"geekgo/week9/webook/internal/repository/dao"
	"time"
)

type ArticleRepository interface {
	Update(ctx context.Context, art domain.Article) error
	Create(ctx context.Context, art domain.Article) (int64, error)
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, aid int64, uid int64, status domain.ArticleStatus) error
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
}

// 制作库和线上库 mysql同库不同表 可以统一使用一个ArticleDAO
type articleRepository struct {
	dao dao.ArticleDAO
}

func (repo *articleRepository) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	da, err := repo.dao.GetPublishedById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return domain.Article{
		Id:      da.Id,
		Title:   da.Title,
		Content: da.Content,
		Author: domain.Author{
			Id: da.AuthorId,
		},
		Status: domain.ArticleStatus(da.Status),
		Ctime:  time.UnixMilli(da.Ctime),
		Utime:  time.UnixMilli(da.Utime),
	}, err
}

func (repo *articleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	da, err := repo.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return domain.Article{
		Id:      da.Id,
		Title:   da.Title,
		Content: da.Content,
		Author: domain.Author{
			Id: da.AuthorId,
		},
		Status: domain.ArticleStatus(da.Status),
		Ctime:  time.UnixMilli(da.Ctime),
		Utime:  time.UnixMilli(da.Utime),
	}, err
}

func (repo *articleRepository) SyncStatus(ctx context.Context, aid int64, uid int64, status domain.ArticleStatus) error {
	return repo.dao.SyncStatus(ctx, aid, uid, status.ToUint8())
}

func (repo *articleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := repo.dao.Sync(ctx, dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
	return id, err
}

func (repo *articleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return repo.dao.Insert(ctx, dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}

func (repo *articleRepository) Update(ctx context.Context, art domain.Article) error {
	return repo.dao.UpdateById(ctx, dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}

func NewArticleRepository(dao dao.ArticleDAO) ArticleRepository {
	return &articleRepository{
		dao: dao,
	}
}
