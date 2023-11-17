package cache

type UserCache interface {
}

type RedisUserCache struct {
}

func NewRedisUserCache() UserCache {
	return &RedisUserCache{}
}
