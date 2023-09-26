package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrCodeSendTooMany   = errors.New("本地验证码发送太频繁")
	ErrCodeNotFound      = errors.New("没有验证码，请先发送验证码")
	ErrCodeVerifyTooMany = errors.New("验证次数太频繁")
)

var (
	ExpireDuration time.Duration = time.Minute * 10
	RetryCnt       int8          = 3
)

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

func NewCodeCache() CodeCache {
	return &LocalCodeCache{
		m: sync.Map{},
	}
}

type LocalCodeCache struct {
	m sync.Map
}

type localCode struct {
	Code           string
	SendTime       time.Time
	ExpireDuration time.Duration
	RetryCnt       int8
}

func (l *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	key := l.key(biz, phone)
	Value, ok := l.m.Load(key)

	if ok {
		if Value.(localCode).SendTime.IsZero() {
			return errors.New("本地验证码缓存出错")
		} else if time.Now().Sub(Value.(localCode).SendTime) < time.Minute {
			return ErrCodeSendTooMany
		} else {
			newValue := localCode{Code: code, SendTime: time.Now(), ExpireDuration: ExpireDuration, RetryCnt: RetryCnt}
			l.m.Swap(key, newValue)
			return nil
		}
	} else {
		newValue := localCode{Code: code, SendTime: time.Now(), ExpireDuration: ExpireDuration, RetryCnt: RetryCnt}
		l.m.Store(key, newValue)
		return nil
	}

}

func (l *LocalCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	key := l.key(biz, phone)
	cacheValue, ok := l.m.Load(key)
	if !ok {
		return false, ErrCodeNotFound
	}

	if cacheValue.(localCode).RetryCnt <= 0 {
		return false, ErrCodeVerifyTooMany
	} else if time.Now().Sub(cacheValue.(localCode).SendTime) >= cacheValue.(localCode).ExpireDuration {
		return false, nil //  验证码过期
	} else if time.Now().Sub(cacheValue.(localCode).SendTime) < cacheValue.(localCode).ExpireDuration {
		if inputCode == cacheValue.(localCode).Code {
			newValue := localCode{Code: cacheValue.(localCode).Code, SendTime: cacheValue.(localCode).SendTime,
				ExpireDuration: cacheValue.(localCode).ExpireDuration, RetryCnt: 0}
			l.m.Swap(key, newValue)
			return true, nil
		} else {
			newValue := localCode{Code: cacheValue.(localCode).Code, SendTime: cacheValue.(localCode).SendTime,
				ExpireDuration: cacheValue.(localCode).ExpireDuration, RetryCnt: cacheValue.(localCode).RetryCnt - int8(1)}
			l.m.Swap(key, newValue)
			return false, nil // 验证码输入错误
		}
	} else {
		return false, errors.New("验证码验证失败，unknown error")
	}
}

func (l *LocalCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("%s:%s", biz, phone)
}
