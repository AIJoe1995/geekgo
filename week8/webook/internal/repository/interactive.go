package repository

import (
	"context"
	"geekgo/week8/webook/internal/domain"
	"geekgo/week8/webook/internal/repository/cache"
	"geekgo/week8/webook/internal/repository/dao"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	IncrCollect(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
	TopNLike(ctx context.Context, biz string, num int64) ([]domain.InteractiveArticle, error)
}

type CachedReadCntRepository struct {
	//cache cache.InteractiveCache
	cache cache.InteractiveCacheV1
	dao   dao.InteractiveDAO
}

func (c *CachedReadCntRepository) TopNLike(ctx context.Context, biz string, num int64) ([]domain.InteractiveArticle, error) {
	topn, err := c.cache.TopNLike(ctx, biz, num)
	if err == nil {
		return topn, nil
	}
	// 查询数据库
	topnDB, err := c.dao.TopNLike(ctx, biz, num)
	topn = []domain.InteractiveArticle{}
	for i := 0; i < len(topnDB); i++ {
		topn = append(topn, domain.InteractiveArticle{
			ArtId:      topnDB[i].BizId,
			LikeCnt:    topnDB[i].LikeCnt,
			ReadCnt:    topnDB[i].ReadCnt,
			CollectCnt: topnDB[i].CollectCnt,
		})
	}
	return topn, err

}

func (c *CachedReadCntRepository) Liked(ctx context.Context, biz string, bizId int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikeInfo(ctx, biz, bizId, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrDataNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedReadCntRepository) Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectInfo(ctx, biz, bizId, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrDataNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedReadCntRepository) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	intr, err := c.cache.Get(ctx, biz, bizId)
	if err == nil {
		return intr, nil
	}
	ie, err := c.dao.Get(ctx, biz, bizId)
	if err == nil {
		res := c.entityToDomain(ie)
		if er := c.cache.Set(ctx, biz, bizId, res); er != nil {
			// 记录日志
		}
		return res, nil
	}
	return domain.Interactive{}, err
}

func (c *CachedReadCntRepository) entityToDomain(intr dao.Interactive) domain.Interactive {
	return domain.Interactive{

		ReadCnt:    intr.ReadCnt,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
	}
}

func (c *CachedReadCntRepository) IncrCollect(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	err := c.dao.InsertCollectInfo(ctx, biz, bizId, cid, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrCollectCntIfPresent(ctx, biz, bizId)
}

func (c *CachedReadCntRepository) IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	// repository调用dao和cache 先更新数据库 后更新缓存， 数据库除了处理like总数计数外，还要处理uid对应biz bizId的like数据记录
	err := c.dao.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrLikeCntIfPresent(ctx, biz, bizId)
}

func (c *CachedReadCntRepository) DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.DecrLikeCntIfPresent(ctx, biz, bizId)
}

func NewCachedReadCntRepository(cache cache.InteractiveCacheV1, dao dao.InteractiveDAO) *CachedReadCntRepository {
	return &CachedReadCntRepository{cache: cache, dao: dao}
}

//func NewCachedReadCntRepository(cache cache.InteractiveCache, dao dao.InteractiveDAO) *CachedReadCntRepository {
//	return &CachedReadCntRepository{cache: cache, dao: dao}
//}

func (c *CachedReadCntRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	//go func() {
	//	c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
	//}()
	//return err

	return c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}
