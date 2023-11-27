package job

import (
	"context"
	"github.com/redis/go-redis/v9"
	"math"
	rand "math/rand"
	"time"
)

const nodesLoad = "nodesLoad"

type Load struct {
	client      redis.Cmdable
	loadChannel chan int64
}

func NewLoad(client redis.Cmdable) *Load {
	return &Load{
		client:      client,
		loadChannel: make(chan int64),
	}
}

func (l *Load) SetLoad(instId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	load := <-l.loadChannel
	_, err := l.client.ZAdd(ctx, nodesLoad, redis.Z{
		Score:  float64(load),
		Member: instId,
	}).Result()
	return err
}

func (l *Load) GetLoadRank(instId string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := l.client.ZRank(ctx, nodesLoad, instId).Result()
	return res, err
}

func (l *Load) generateLoad() {
	max_num := math.MaxInt64

	LoadTicker := time.NewTicker(time.Minute)
	defer LoadTicker.Stop()
	for {
		select {
		case <-LoadTicker.C:
			rand := rand.Int63n(int64(max_num))
			l.loadChannel <- rand
			// 如果loadChannel 没有位置接收 会怎么样 会阻塞吗？
		}
	}

}
