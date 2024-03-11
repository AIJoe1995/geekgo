package service

import (
	"context"
	"geekgo/week17/account/domain"
	"geekgo/week17/account/repository"
)

type AccountService interface {
	Credit(ctx context.Context, cr domain.Credit) error
}
type accountService struct {
	repo repository.AccountRepository
}

func (a *accountService) Credit(ctx context.Context, cr domain.Credit) error {
	return a.repo.AddCredit(ctx, cr)
}

func NewAccountService(repo repository.AccountRepository) AccountService {
	return &accountService{repo: repo}
}
