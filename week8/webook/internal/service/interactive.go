package service

import (
	"context"
	"geekgo/week8/webook/internal/domain"
	"geekgo/week8/webook/internal/repository"
	"golang.org/x/sync/errgroup"
)

// 提供阅读计数 点赞 收藏等服务 articleHandler 聚合articleService和interactiveService

type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, bizId int64, uid int64) error
	CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error
	Collect(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error)
	TopNLike(ctx context.Context, biz string, num int64) ([]domain.InteractiveArticle, error)
}

type interactiveService struct {
	repo repository.InteractiveRepository
}

func (i *interactiveService) TopNLike(ctx context.Context, biz string, num int64) ([]domain.InteractiveArticle, error) {
	topn, err := i.repo.TopNLike(ctx, biz, num)
	if err != nil {
		return []domain.InteractiveArticle{}, err
	}
	return topn, err

}

func (i *interactiveService) Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error) {
	intr, err := i.repo.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		intr.Liked, err = i.repo.Liked(ctx, biz, bizId, uid)
		return err
	})
	eg.Go(func() error {
		intr.Collected, err = i.repo.Collected(ctx, biz, bizId, uid)
		return err
	})
	err = eg.Wait()
	if err != nil {
		// 记录日志
	}
	return intr, err

}

func (i *interactiveService) Collect(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	// collect 要把内容收藏到对应收藏夹 并且collectCnt + 1
	return i.repo.IncrCollect(ctx, biz, bizId, cid, uid)
}

func (i *interactiveService) Like(ctx context.Context, biz string, bizId int64, uid int64) error {
	// 与阅读计数类似，like也要维护一篇文章对应的like计数， 但除此之外，还要展示每个用户是否like改文章，需要增加一个数据库表 存储这个信息。
	return i.repo.IncrLike(ctx, biz, bizId, uid)
}

func (i *interactiveService) CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	return i.repo.DecrLike(ctx, biz, bizId, uid)
}

func NewInteractiveService(repo repository.InteractiveRepository) *interactiveService {
	return &interactiveService{repo: repo}

}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}
