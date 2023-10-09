package repository

import (
	"context"
	"geekgo/week5/webook/repository/cache"
)

var (
	ErrCodeSendTooMany        = cache.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
)

type CodeRepository interface {
	Store(ctx context.Context, biz string, phone string, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type CachedCodeRepository struct {
	cache cache.CodeCache
}

func (repo CachedCodeRepository) Store(ctx context.Context, biz string, phone string, code string) error {
	// 调用cache.Set
	return repo.cache.Set(ctx, biz, phone, code)
}

func (repo CachedCodeRepository) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	// 调用cache.Verify
	return repo.cache.Verify(ctx, biz, phone, inputCode)
}

func NewCodeRepository(c cache.CodeCache) CodeRepository {
	return &CachedCodeRepository{
		cache: c,
	}
}
