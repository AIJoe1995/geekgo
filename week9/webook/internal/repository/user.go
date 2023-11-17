package repository

import (
	"context"
	"database/sql"
	"geekgo/week9/webook/internal/domain"
	"geekgo/week9/webook/internal/repository/cache"
	"geekgo/week9/webook/internal/repository/dao"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrDataNotFound
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
}

type userRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &userRepository{
		dao:   dao,
		cache: cache,
	}
}

func (repo *userRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	ud, err := repo.dao.FindByEmail(ctx, email)
	return repo.entityToDomain(ud), err
}

func (repo *userRepository) Create(ctx context.Context, user domain.User) error {
	err := repo.dao.Insert(ctx, dao.User{
		Id: user.Id,
		Email: sql.NullString{
			String: user.Email,
			Valid:  user.Email != "",
		},
		Password: user.Password,
	})
	return err

}

func (repo *userRepository) entityToDomain(ud dao.User) domain.User {
	return domain.User{
		Id:       ud.Id,
		Email:    ud.Email.String,
		Password: ud.Password,
	}
}
