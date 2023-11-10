package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"geekgo/week8/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"sort"
	"strconv"
	"time"
)

// 分成n个ordered set装对应biz bizId 的like read collect
// 具体实现 Set Get时 需要先 % 出放在哪个里面
type InteractiveCacheV1 interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error
	TopNLike(ctx context.Context, biz string, num int64) ([]domain.InteractiveArticle, error)
}

type interactiveCacheV1 struct {
	client     redis.Cmdable
	expiration time.Duration
	batchNum   int64
}

var (
	//go:embed lua/interative_incr_cnt_zs.lua
	luaIncrCntZS string
)

var (
	//go:embed lua/interative_topn_zs.lua
	luaTopNZS string
)

const (
	likeCntZSKey    = "like_cnt"
	collectCntZSKey = "collect_cnt"
	readCntZSKey    = "read_cnt"
)

func NewInteractiveCacheV1(client redis.Cmdable) InteractiveCacheV1 {
	return &interactiveCacheV1{client: client, expiration: time.Minute * 30, batchNum: int64(10)}
}

func (i *interactiveCacheV1) key(biz string, field string) string {
	return fmt.Sprintf("interactive:%s:%s", biz, field)
}

func (i *interactiveCacheV1) batchKey(redisKey string, bizId int64) string {
	batchNo := bizId % i.batchNum
	return fmt.Sprintf("%s:%d", redisKey, batchNo)
}

func (i *interactiveCacheV1) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	redisKey := i.key(biz, readCntZSKey)
	batchKey := i.batchKey(redisKey, bizId)
	// - lua脚本有没有问题 这样执行对吗 参数？
	return i.client.Eval(ctx, luaIncrCntZS, []string{batchKey}, bizId, 1).Err()
}

func (i *interactiveCacheV1) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	redisKey := i.key(biz, likeCntZSKey)
	batchKey := i.batchKey(redisKey, bizId)
	// - lua脚本有没有问题 这样执行对吗 参数？
	return i.client.Eval(ctx, luaIncrCntZS, []string{batchKey}, bizId, 1).Err()
}

func (i *interactiveCacheV1) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	redisKey := i.key(biz, likeCntZSKey)
	batchKey := i.batchKey(redisKey, bizId)
	// - lua脚本有没有问题 这样执行对吗 参数？
	return i.client.Eval(ctx, luaIncrCntZS, []string{batchKey}, bizId, -1).Err()
}

func (i *interactiveCacheV1) IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	redisKey := i.key(biz, collectCntZSKey)
	batchKey := i.batchKey(redisKey, bizId)
	// - lua脚本有没有问题 这样执行对吗 参数？
	return i.client.Eval(ctx, luaIncrCntZS, []string{batchKey}, bizId, 1).Err()
}

func (i *interactiveCacheV1) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	redisLikeKey := i.key(biz, likeCntZSKey)
	batchLikeKey := i.batchKey(redisLikeKey, bizId%i.batchNum)
	redisReadKey := i.key(biz, readCntZSKey)
	batchReadKey := i.batchKey(redisReadKey, bizId%i.batchNum)
	redisCollectKey := i.key(biz, collectCntZSKey)
	batchCollectKey := i.batchKey(redisCollectKey, bizId%i.batchNum)

	mem := fmt.Sprintf("%d", bizId)
	// 写一个lua统一获取？

	like_cnt, err1 := i.client.ZScore(ctx, batchLikeKey, mem).Result()
	collect_cnt, err2 := i.client.ZScore(ctx, batchCollectKey, mem).Result()
	read_cnt, err3 := i.client.ZScore(ctx, batchReadKey, mem).Result()
	if err1 == nil && err2 == nil && err3 == nil {
		return domain.Interactive{
			LikeCnt:    int64(like_cnt),
			CollectCnt: int64(collect_cnt),
			ReadCnt:    int64(read_cnt),
		}, nil
	}
	return domain.Interactive{}, errors.New("redis cnt info error")
}

func (i *interactiveCacheV1) Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error {
	redisLikeKey := i.key(biz, likeCntZSKey)
	batchLikeKey := i.batchKey(redisLikeKey, bizId%i.batchNum)
	redisReadKey := i.key(biz, readCntZSKey)
	batchReadKey := i.batchKey(redisReadKey, bizId%i.batchNum)
	redisCollectKey := i.key(biz, collectCntZSKey)
	batchCollectKey := i.batchKey(redisCollectKey, bizId%i.batchNum)

	mem := fmt.Sprintf("%d", bizId)
	_, err1 := i.client.ZAdd(ctx, batchLikeKey, redis.Z{Score: float64(intr.LikeCnt), Member: mem}).Result()
	_, err2 := i.client.ZAdd(ctx, batchReadKey, redis.Z{Score: float64(intr.ReadCnt), Member: mem}).Result()
	_, err3 := i.client.ZAdd(ctx, batchCollectKey, redis.Z{Score: float64(intr.CollectCnt), Member: mem}).Result()

	if err1 != nil || err2 != nil || err3 != nil {
		return errors.New("redis cnt set error")
	}
	return nil

}

func (i *interactiveCacheV1) TopNLike(ctx context.Context, biz string, num int64) ([]domain.InteractiveArticle, error) {
	// 这里除了返回artid 之外 还要返回likecnt 以及collectcnt readcnt
	like_key := i.key(biz, likeCntZSKey)
	collect_key := i.key(biz, collectCntZSKey)
	read_key := i.key(biz, readCntZSKey)
	all_record := []domain.InteractiveArticle{} // 保存topN bizId
	for cnt := int64(0); cnt < i.batchNum; cnt++ {
		batch_like_key := fmt.Sprintf("%s:%d", like_key, cnt)
		batch_collect_key := fmt.Sprintf("%s:%d", collect_key, cnt)
		batch_read_key := fmt.Sprintf("%s:%d", read_key, cnt)
		// 这里参数可能不对
		res, err := i.client.Eval(ctx, luaTopNZS, []string{batch_like_key}, num).Result()

		if err != nil {
			return []domain.InteractiveArticle{}, err
		}
		// 上面拿到的是bizId string, ZMSCore 获取对应的分数即like_cnt
		res_i := res.([]interface{})
		keyId_list := []string{}
		for _, keyId := range res_i {
			keyId_list = append(keyId_list, keyId.(string))
		}
		like_cnts := i.client.ZMScore(ctx, batch_like_key, keyId_list...).Val()
		collect_cnts := i.client.ZMScore(ctx, batch_collect_key, keyId_list...).Val()
		read_cnts := i.client.ZMScore(ctx, batch_read_key, keyId_list...).Val()

		for idx, keyId := range keyId_list {
			keyId_i, err := strconv.Atoi(keyId)
			if err != nil {
				//
			}
			all_record = append(all_record, domain.InteractiveArticle{
				ArtId:      int64(keyId_i),
				LikeCnt:    int64(like_cnts[idx]),
				CollectCnt: int64(collect_cnts[idx]),
				ReadCnt:    int64(read_cnts[idx]),
			})
		}
		// 遍历返回的topN结果 拿到bizId 和 score 保存到all_record
	}
	// 对all_record 排序 得到最终的topN
	//topn := all_record
	sort.Slice(all_record, func(i, j int) bool {
		return all_record[i].LikeCnt > all_record[j].LikeCnt
	})
	if len(all_record) < int(num) {
		return all_record, nil
	}
	return all_record[:num], nil
}

func (i *interactiveCacheV1) TopNCollect(ctx context.Context, biz string) error {
	//TODO implement me
	panic("implement me")
}

func (i *interactiveCacheV1) TopNRead(ctx context.Context, biz string) error {
	//TODO implement me
	panic("implement me")
}

var (
	//go:embed lua/interative_incr_cnt.lua
	luaIncrCnt string
)

const (
	fieldReadCnt    = "read_cnt"
	fieldCollectCnt = "collect_cnt"
	fieldLikeCnt    = "like_cnt"
)

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error
}

type RedisInteractiveCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func (r *RedisInteractiveCache) Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error {

	key := r.key(biz, bizId)
	err := r.client.HMSet(ctx, key,
		fieldCollectCnt, intr.CollectCnt,
		fieldLikeCnt, intr.LikeCnt,
		fieldReadCnt, intr.ReadCnt,
	).Err()
	if err != nil {
		return err
	}
	return r.client.Expire(ctx, key, time.Minute*15).Err()

}

func (r *RedisInteractiveCache) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	data, err := r.client.HGetAll(ctx, r.key(biz, bizId)).Result()
	if err != nil {
		return domain.Interactive{}, err
	}
	if len(data) == 0 {
		return domain.Interactive{}, ErrKeyNotExist
	}

	// 理论上来说，这里不可能有 error
	collectCnt, _ := strconv.ParseInt(data[fieldCollectCnt], 10, 64)
	likeCnt, _ := strconv.ParseInt(data[fieldLikeCnt], 10, 64)
	readCnt, _ := strconv.ParseInt(data[fieldReadCnt], 10, 64)

	return domain.Interactive{
		// 懒惰的写法
		CollectCnt: collectCnt,
		LikeCnt:    likeCnt,
		ReadCnt:    readCnt,
	}, err
}

func (r *RedisInteractiveCache) IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		// read_cnt +1
		fieldCollectCnt, 1).Err()
}

func (r *RedisInteractiveCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		// read_cnt +1
		fieldLikeCnt, -1).Err()
}

func (r *RedisInteractiveCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		// read_cnt +1
		fieldLikeCnt, 1).Err()
}

func NewRedisInteractiveCache(client redis.Cmdable) *RedisInteractiveCache {
	return &RedisInteractiveCache{client: client}

}

func (r *RedisInteractiveCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		// read_cnt +1
		fieldReadCnt, 1).Err()
}

func (r *RedisInteractiveCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
