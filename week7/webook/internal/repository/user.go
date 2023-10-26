package repository

import (
	"context"
	"database/sql"
	"geekgo/week7/webook/internal/domain"
	"geekgo/week7/webook/internal/repository/dao"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	// 调用dao的create 这里需要把domain.User转换成dao.User
	return r.dao.Insert(ctx, r.domainToEntity(u))

}
func (r *UserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			// 我确实有手机号
			Valid: u.Email != "",
		},
		Password: u.Password,
	}
}

func (r *UserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Password: u.Password,
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	return r.entityToDomain(u), err
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{dao: dao}

}
