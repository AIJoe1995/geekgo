package cache

var _ RewardCache = (*RedisRewardCache)(nil)

type RewardCache interface {
}

type RedisRewardCache struct {
}
