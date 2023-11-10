package service

import (
	"context"
	"geekgo/week8/webook/internal/domain"
	"geekgo/week8/webook/internal/repository/article"
)

type ArticleService struct {
	repo *article.ArticleRepository

	// 在service层面聚合 publish到制作库和线上库
	authorRepo *article.ArticleAuthorRepository
	readerRepo *article.ArticleReaderRepository
}

func NewArticleServiceV1(authorRepo *article.ArticleAuthorRepository, readerRepo *article.ArticleReaderRepository) *ArticleService {
	return &ArticleService{
		authorRepo: authorRepo,
		readerRepo: readerRepo,
	}
}

func (svc *ArticleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	//  authorRepo操作制作库
	// readerRepo操作线上库
	id := art.Id
	var err error
	if art.Id == 0 {
		id, err = svc.authorRepo.Create(ctx, art) // 新建
	} else {
		err = svc.authorRepo.Update(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	// readerRepo操作到线上库
	for i := 0; i < 3; i++ {
		err = svc.readerRepo.Save(ctx, art)
		if err == nil {
			break
		}
		// 记录日志
		// 接入metrics tracing
	}
	if err != nil {
		// 保存到线上库重试失败
		return 0, err
	}

	return id, nil

}

func (svc *ArticleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	// 调用repository保存到制作库
	return svc.repo.Save(ctx, art)
}

func (svc *ArticleService) Publish(ctx context.Context, art domain.Article) (int64, error) {

	return svc.repo.Publish(ctx, art)
}

func (svc *ArticleService) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	return svc.repo.GetPublishedById(ctx, id)
}

func (svc *ArticleService) Withdraw(ctx context.Context, art domain.Article) error {
	// withdraw 找到这篇文章 修改 文章状态
	return svc.repo.SyncStatus(ctx, art, domain.ArticleStatusPrivate.ToUint8())
}

func (svc *ArticleService) GetPublishedByIds(ctx context.Context, ids []int64) ([]domain.Article, error) {
	return svc.repo.GetPublishedByIds(ctx, ids)
}

func NewArticleService(repo *article.ArticleRepository) *ArticleService {
	return &ArticleService{repo: repo}
}
