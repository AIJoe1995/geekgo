package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"testing"
)

func TestPrepareData(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	//likeCntZSKey like_cnt
	biz := "article"
	key := fmt.Sprintf("interactive:%s:%s", biz, "like_cnt")

	// 向redis中插入数据
	for i := 1; i <= 100; i++ {
		bizId := i
		like_cnt := i * 2
		batchNo := i % 10
		batchKey := fmt.Sprintf("%s:%d", key, batchNo)
		client.ZAdd(
			context.Background(),
			batchKey,
			redis.Z{Score: float64(like_cnt),
				Member: bizId,
			},
		)
	}
}

func TestDeleteRedisData(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	//likeCntZSKey like_cnt
	biz := "article"
	key := fmt.Sprintf("interactive:%s:%s", biz, "like_cnt")
	for i := 0; i < 10; i++ {
		batchKey := fmt.Sprintf("%s:%d", key, i)
		client.Del(context.Background(), batchKey)
	}
}
