package service

import (
	"context"
	"geekgo/week5/webook/domain"
	"geekgo/week5/webook/repository"
)

type UserService interface {
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func (svc userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	user, err := svc.repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		return user, err
	}

	user = domain.User{
		Phone: phone,
	}
	err = svc.repo.Create(ctx, user)
	if err != nil && err != repository.ErrUserDuplicate {
		return user, err
	}
	return svc.repo.FindByPhone(ctx, phone)
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}
