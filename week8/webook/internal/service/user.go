package service

import (
	"context"
	"geekgo/week8/webook/internal/domain"
	"geekgo/week8/webook/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserDuplicate = repository.ErrUserDuplicate
var ErrInvalidUserOrPassword = repository.ErrInvalidUserOrPassword
var ErrUserNotFound = repository.ErrUserNotFound

type UserService struct {
	repo *repository.UserRepository
}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)

	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if err == ErrUserNotFound {
		return domain.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, err
}

func (svc *UserService) Profile(ctx context.Context, uid int64) (domain.User, error) {
	return svc.repo.FindById(ctx, uid)
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}
