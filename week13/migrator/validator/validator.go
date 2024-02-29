package validator

import (
	"context"
	"geekgo/week13/migrator"
	"geekgo/week13/migrator/events"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
)

// 校验数据 分成两种情况 从源表向目标表的校验 从源表取数据判断目标表是否存在 因为可能有在目标表但不在源表的 所以目标表到源表也需要校验一次 从目标表取数据 看源表存不存在

// 迁移过程 会出现以目标表为准和源表为准

// 从表中取数据 可以有全量数据和增量数据两种 增量数据使用utime大于给定值来进行validate

type Validator[T migrator.Entity] struct {
	base      *gorm.DB
	target    *gorm.DB
	direction string
	utime     int64

	// 批量校验
	batchSize int

	// 没有数据的时候 增量校验不能直接停止 睡眠一会 然后再看有没有新增数据
	sleepInterval time.Duration

	// 对比的结果放到消息中间件
	producer events.Producer
}

// NewValidator 所有实现了migrator.Entity接口(ID和CompareTo)的表都可以进行validate
func NewValidator[T migrator.Entity](base *gorm.DB,
	target *gorm.DB,
	direction string,
	producer events.Producer,
) *Validator[T] {
	return &Validator[T]{
		base:          base,
		target:        target,
		direction:     direction,
		producer:      producer,
		batchSize:     100,
		sleepInterval: 0,
	}
}

func (v *Validator[T]) Utime(utime int64) *Validator[T] {
	v.utime = utime
	return v
}

func (v *Validator[T]) SleepInterval(i time.Duration) *Validator[T] {
	v.sleepInterval = i
	return v
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group
	eg.Go(func() error {
		// baseToTarget
		return v.baseToTarget(ctx)
	})
	eg.Go(func() error {
		// targetToBase
		return v.targetToBase(ctx)
	})
	return eg.Wait()
}

func (v *Validator[T]) baseToTarget(ctx context.Context) error {
	offset := 0
	for {
		var srcs []T
		dbCtx, cancel := context.WithTimeout(ctx, time.Second*2)
		// 按照id排序 offset 如果插入的id不是越来越大的 全量检查可能会漏掉记录
		//后插入的id utime保证一定更大这样miss的记录可以在增量检查中做validate
		err := v.base.WithContext(dbCtx).Order("id").
			Where("utime >= ?", v.utime).
			Offset(offset).Limit(v.batchSize).Find(&srcs).Error
		cancel()
		switch err {
		case gorm.ErrRecordNotFound:
			// 没有记录了 如果是增量检查 就睡眠 全量检查应该可以直接返回
			if v.sleepInterval > 0 {
				time.Sleep(v.sleepInterval)
				continue
			} else {
				return nil
			}
		case context.DeadlineExceeded:
			continue
		case context.Canceled:
			// 超时应该重试吧 Cancled应该是直接return吧
			return nil
		case nil:
			// 没有错误 把对应的数据从target取出来
			// 这个函数里分成几种情况
			// 1.  base里有 target没有
			// 2. base和target不同
			// 3. base和target相同
			v.dstDiff(ctx, srcs)
		default:
			// 记录日志
		}
		offset += len(srcs)

	}
}

func (v *Validator[T]) targetToBase(ctx context.Context) error {
	offset := 0
	for {
		var dsts []T
		dbCtx, cancel := context.WithTimeout(ctx, time.Second*2)
		// 按照id排序 offset 如果插入的id不是越来越大的 全量检查可能会漏掉记录
		//后插入的id utime保证一定更大这样miss的记录可以在增量检查中做validate
		err := v.target.WithContext(dbCtx).Order("id").
			Where("utime >= ?", v.utime).
			Offset(offset).Limit(v.batchSize).Find(&dsts).Error
		cancel()
		switch err {
		case gorm.ErrRecordNotFound:
			// 没有记录了 如果是增量检查 就睡眠 全量检查应该可以直接返回
			if v.sleepInterval > 0 {
				time.Sleep(v.sleepInterval)
				continue
			} else {
				return nil
			}
		case context.DeadlineExceeded:
			continue
		case context.Canceled:
			// 超时应该重试吧 Cancled应该是直接return吧
			return nil
		case nil:
			// 没有错误 把对应的数据从target取出来
			// 这个函数里分成几种情况
			// 1.  base里有 target没有
			// 2. base和target不同
			// 3. base和target相同
			v.missingTarget(ctx, dsts)
		default:
			// 记录日志
		}
		offset += len(dsts)

	}
}

func (v *Validator[T]) dstDiff(ctx context.Context, srcs []T) {
	srcs_map := make(map[int64]T)
	dsts_map := make(map[int64]T)
	ids := make([]int64, len(srcs))
	for _, src := range srcs {
		srcs_map[src.ID()] = src
		ids = append(ids, src.ID())
	}
	var dsts []T
	dbCtx, cancel := context.WithTimeout(ctx, time.Second*2)
	err := v.target.WithContext(dbCtx).Where("id IN ?", ids).Find(&dsts).Error
	cancel()
	switch err {
	case gorm.ErrRecordNotFound:
		// 没有记录 都是missing
	case nil:
		// 之后对比记录
	default:
		// 记录日志 default出现数据库错误 应该重新validate
	}
	for _, dst := range dsts {
		dsts_map[dst.ID()] = dst
	}
	for id, src := range srcs_map {
		dst, ok := dsts_map[id]
		if !ok {
			// report missing
			v.notify(id, events.InconsistentEventTypeTargetMissing)

		} else {
			if !src.CompareTo(dst) {
				// report mismatch
				v.notify(id, events.InconsistentEventTypeNotEqual)

			}
		}

	}
}

func (v *Validator[T]) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	evt := events.InconsistentEvent{
		Direction: v.direction,
		ID:        id,
		Type:      typ,
	}
	err := v.producer.ProduceInconsistentEvent(ctx, evt)
	if err != nil {
		// 记录日志
	}

}

func (v *Validator[T]) missingTarget(ctx context.Context, dsts []T) {
	ids := make([]int64, len(dsts))
	srcs_map := make(map[int64]T)
	dsts_map := make(map[int64]T)

	for _, dst := range dsts {
		dsts_map[dst.ID()] = dst
		ids = append(ids, dst.ID())
	}
	var srcs []T
	dbCtx, cancel := context.WithTimeout(ctx, time.Second*2)
	err := v.base.WithContext(dbCtx).Where("id IN ?", ids).Find(&srcs).Error
	cancel()
	switch err {
	case gorm.ErrRecordNotFound:
		// 没有记录 都是missing
	case nil:
		// 之后对比记录
	default:
		// 记录日志 default出现数据库错误 应该重新validate
	}

	for _, src := range srcs {
		srcs_map[src.ID()] = src
	}
	for id, _ := range dsts_map {
		_, ok := srcs_map[id]
		if !ok {
			// report missing
			v.notify(id, events.InconsistentEventTypeBaseMissing)
		}
	}
}