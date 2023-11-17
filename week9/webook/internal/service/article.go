package service

import (
	"context"
	"geekgo/week9/webook/internal/domain"
	events "geekgo/week9/webook/internal/events/article"
	"geekgo/week9/webook/internal/repository"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	Withdraw(ctx context.Context, aid int64, uid int64) error
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	producer events.Producer
}

func (svc *articleService) GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error) {
	art, err := svc.repo.GetPublishedById(ctx, id)
	if err == nil {
		go func() {
			er := svc.producer.ProduceReadEvent(ctx,
				events.ReadEvent{
					Uid: uid,
					Aid: id,
				})
			if er != nil {
				// 记录日志
			}
		}()
	}
	return art, err
}

func (svc *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return svc.repo.GetById(ctx, id)
}

func (svc *articleService) Withdraw(ctx context.Context, aid int64, uid int64) error {
	status := domain.ArticleStatusPrivate
	return svc.repo.SyncStatus(ctx, aid, uid, status)
}

func (svc *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	return svc.repo.Sync(ctx, art)
}

func (svc *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	// 更新
	if art.Id > 0 {
		err := svc.repo.Update(ctx, art)
		return art.Id, err
	}
	// 创建
	return svc.repo.Create(ctx, art)
}

func NewArticleService(repo repository.ArticleRepository, producer events.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
	}
}
