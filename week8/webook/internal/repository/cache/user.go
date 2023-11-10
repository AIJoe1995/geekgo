package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"geekgo/week8/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

// repository里包括dao和cache 在repository层面组合dao cache 现在缓存中查询， 缓存中没有再从数据库查询

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
}

type RedisUserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewRedisUserCache(client redis.Cmdable) *RedisUserCache {
	return &RedisUserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

func (cache *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	// client.Get返回*StringCmd
	// StringCmd.Bytes 返回key对应value的[]byte形式和err 也可以StringCmd.Result
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

func (cache *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.key(u.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
