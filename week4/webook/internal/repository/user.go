package repository

import (
	"context"
	"database/sql"
	"geekgo/week4/webook/internal/domain"
	"geekgo/week4/webook/internal/repository/dao"
	"time"
)

var (
	ErrUserNotFound  = dao.ErrUserNotFound
	ErrUserDuplicate = dao.ErrUserDuplicate
)

type UserRepository interface {
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	Create(ctx context.Context, u domain.User) error
}

type userRepository struct {
	dao dao.UserDAO
}

func (up *userRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := up.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return up.entityToDomain(u), nil

}

func NewUserRepository(dao dao.UserDAO) UserRepository {
	return &userRepository{
		dao: dao,
	}
}

func (up *userRepository) Create(ctx context.Context, u domain.User) error {
	return up.dao.Insert(ctx, up.domainToEntity(u))
}

func (up *userRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			// 我确实有手机号
			Valid: u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,

		Ctime: u.Ctime.UnixMilli(),
	}
}

func (up *userRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Password: u.Password,
		Phone:    u.Phone.String,

		Ctime: time.UnixMilli(u.Ctime),
	}
}
