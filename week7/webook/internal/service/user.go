package service

import (
	"context"
	"geekgo/week7/webook/internal/domain"
	"geekgo/week7/webook/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicate = repository.ErrUserDuplicate
	ErrUserNotFound  = repository.ErrUserNotFound
)

type UserService struct {
	repo *repository.UserRepository
}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	//svc层对密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	// FindByEmail
	user, err := svc.repo.FindByEmail(ctx, email)
	return user, err

}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}
