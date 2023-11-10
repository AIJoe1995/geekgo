package repository

import (
	"context"
	"database/sql"
	"errors"
	"geekgo/week8/webook/internal/domain"
	"geekgo/week8/webook/internal/repository/cache"
	"geekgo/week8/webook/internal/repository/dao"
)

var ErrUserDuplicate = dao.ErrUserDuplicate
var ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")
var ErrUserNotFound = dao.ErrUserNotFound

type UserRepository struct {
	dao   *dao.UserDAO
	cache cache.UserCache
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func NewUserRepository(dao *dao.UserDAO, cache cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *UserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Password: u.Password,
		Nickname: u.Nickname,
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {

	u, err := r.dao.FindByEmail(ctx, email)

	return r.entityToDomain(u), err
}

func (r *UserRepository) entityToDomain(user dao.User) domain.User {
	return domain.User{
		Id:       user.Id,
		Email:    user.Email.String,
		Password: user.Password,
		Nickname: user.Nickname,
	}
}

func (r *UserRepository) FindById(ctx context.Context, uid int64) (domain.User, error) {
	// 增加逻辑 先从缓存查
	u, err := r.cache.Get(ctx, uid)
	if err == nil {
		return u, nil
	}

	userdao, err := r.dao.FindById(ctx, uid)
	if err != nil {
		return domain.User{}, err
	}
	u = r.entityToDomain(userdao)

	// 开启一个goroutine 设置缓存
	go func() {
		_ = r.cache.Set(ctx, u)
	}()

	return u, err
}
