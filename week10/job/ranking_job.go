package job

import (
	"context"
	"github.com/google/uuid"
	rlock "github.com/gotomicro/redis-lock"
	"sync"
	"time"
)

type RankingJob struct {
	NodeId    string
	key       string
	timeout   time.Duration
	localLock *sync.Mutex
	lock      *rlock.Lock
	client    *rlock.Client

	load *Load
}

func NewRankingJob(client *rlock.Client, load *Load) *RankingJob {
	return &RankingJob{
		NodeId: uuid.New().String(),
		client: client,
		load:   load,
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	// 本地上锁
	r.localLock.Lock()
	defer r.localLock.Unlock()
	if r.lock == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      0,
		}, time.Second)
		if err != nil {
			return nil
		}
		r.load.SetLoad(r.NodeId)
		// 要判断是否结点的load排名在靠前位置 不在的话不能拿锁 直接return
		loadRank, err := r.load.GetLoadRank(r.NodeId)
		if loadRank >= 3 {
			return nil
		}
		r.lock = lock
		go func() {
			r.localLock.Lock()
			defer r.localLock.Unlock()
			loadRank, err = r.load.GetLoadRank(r.NodeId)
			if err != nil {
				//
			}
			if loadRank < 3 {
				err1 := lock.AutoRefresh(r.timeout/2, time.Second)
				// 这里说明退出了续约机制
				// 续约失败了怎么办？
				if err1 != nil {
					// 不怎么办
					// 争取下一次，继续抢锁
					//r.l.Error("续约失败", logger.Error(err))
				}
				//r.lock = nil 这里为什么r.lock=nil?
			}
		}()
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return topn(ctx)
}

func topn(ctx context.Context) error {
	return nil
}
