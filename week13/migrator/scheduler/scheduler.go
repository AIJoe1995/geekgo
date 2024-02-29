package scheduler

import (
	"context"
	"fmt"
	"geekgo/week13/connpool"
	"geekgo/week13/migrator"
	"geekgo/week13/migrator/events"
	"geekgo/week13/migrator/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"sync"
	"time"
)

// 调度 数据迁移的调度 需要进行 全量校验 增量校验
// 双写 为了不改动业务代码 引入connpool
type Scheduler[T migrator.Entity] struct {
	src *gorm.DB
	dst *gorm.DB

	lock       sync.Mutex
	pattern    string
	cancelFull func()
	cancelIncr func()
	producer   events.Producer

	pool *connpool.DoubleWritePool
}

func NewScheduler[T migrator.Entity](src *gorm.DB, dst *gorm.DB,
	pattern string, cancelFull func(),
	cancelIntr func(), producer events.Producer) *Scheduler[T] {
	return &Scheduler[T]{src: src, dst: dst, cancelFull: cancelFull,
		cancelIncr: cancelIntr, producer: producer,
		pattern: connpool.PatternSrcFirst,
	}
}

func (s *Scheduler[T]) RegisterRoutes(server *gin.RouterGroup) {
	// 将这个暴露为 HTTP 接口
	// 你可以配上对应的 UI
	server.POST("/src_only", s.SrcOnly)
	server.POST("/src_first", s.SrcFirst)
	server.POST("/dst_first", s.DstFirst)
	server.POST("/dst_only", s.DstOnly)
	server.POST("/full/start", s.StartFullValidation)
	server.POST("/full/stop", s.StopFullValidation)
	server.POST("/incr/stop", s.StopIncrementValidation)
	server.POST("/incr/start", s.StartIncrementValidation)
}

func (s *Scheduler[T]) SrcOnly(ctx *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcOnly
	s.pool.ChangePattern(connpool.PatternSrcOnly)
	ctx.JSON(http.StatusOK, Result{
		Msg: "ok",
	})
}

func (s *Scheduler[T]) SrcFirst(ctx *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcFirst
	s.pool.ChangePattern(connpool.PatternSrcFirst)
	ctx.JSON(http.StatusOK, Result{
		Msg: "ok",
	})
}

func (s *Scheduler[T]) DstFirst(ctx *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstFirst
	s.pool.ChangePattern(connpool.PatternDstFirst)
	ctx.JSON(http.StatusOK, Result{
		Msg: "ok",
	})
}

func (s *Scheduler[T]) DstOnly(ctx *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstOnly
	s.pool.ChangePattern(connpool.PatternDstOnly)
	ctx.JSON(http.StatusOK, Result{
		Msg: "ok",
	})
}

func (s *Scheduler[T]) newValidator() (*validator.Validator[T], error) {
	switch s.pattern {
	case connpool.PatternSrcFirst, connpool.PatternSrcOnly:
		return validator.NewValidator[T](s.src, s.dst, "SRC", s.producer), nil
	case connpool.PatternDstFirst, connpool.PatternDstOnly:
		return validator.NewValidator[T](s.dst, s.src, "DST", s.producer), nil
	default:
		return nil, fmt.Errorf("未知的 pattern %s", s.pattern)
	}
}

type Result struct {
	Code int
	Msg  string
	Data any
}

// StartFullValidation 全量校验 需要Validator[T] 需要能够取消 使用ctx ctx.Err 取消ctx来取消全量校验 退出for循环
func (s *Scheduler[T]) StartFullValidation(ctx *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, err := s.newValidator()
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "全量校验出错",
		})
	}
	cancel := s.cancelFull
	var vctx context.Context
	vctx, s.cancelFull = context.WithCancel(context.Background())

	go func() {
		// 先取消上一次的
		cancel()
		err := v.Validate(vctx)
		if err != nil {
			// 记录日志
		}
	}()
	ctx.JSON(http.StatusOK, Result{
		Msg: "ok",
	})

}

func (s *Scheduler[T]) StopFullValidation(ctx *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelFull()
	ctx.JSON(http.StatusOK, Result{
		Msg: "ok",
	})
}

func (s *Scheduler[T]) StopIncrementValidation(ctx *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelIncr()
	ctx.JSON(http.StatusOK, Result{
		Msg: "ok",
	})
}

// StartIncrementValidation 增量校验 需要在正常初始化validator的基础上 设置sleepInterval和utime
func (s *Scheduler[T]) StartIncrementValidation(ctx *gin.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()
	cancel := s.cancelIncr
	type IncrReq struct {
		Utime         int64 `json:"utime"`
		SleepInterval int64 `json:"sleep_interval"`
	}
	var req IncrReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	v, err := s.newValidator()
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "增量校验出错",
		})
	}
	v.SleepInterval(time.Duration(req.SleepInterval) * time.Millisecond).Utime(req.Utime)
	var vctx context.Context
	vctx, s.cancelIncr = context.WithCancel(context.Background())

	go func() {
		cancel()
		err := v.Validate(vctx)
		if err != nil {
			// 记录日志
		}
	}()
	ctx.JSON(http.StatusOK, Result{
		Msg: "ok",
	})
}
