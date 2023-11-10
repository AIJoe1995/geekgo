package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicate = errors.New("邮箱冲突")
	ErrUserNotFound  = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func (d *UserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := d.db.Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			// 邮箱冲突 or 手机号码冲突
			return ErrUserDuplicate
		}
	}
	return err
}

func (d *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := d.db.WithContext(ctx).Where("email=?", email).First(&u).Error
	if err == gorm.ErrRecordNotFound {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}
	return u, err
}

func (d *UserDAO) FindById(ctx context.Context, uid int64) (User, error) {
	var u User
	err := d.db.WithContext(ctx).Where("id=?", uid).First(u).Error
	return u, err
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

type User struct {
	Id       int64          `gorm:"primaryKey,autoIncrement"`
	Email    sql.NullString `gorm:"unique"`
	Password string
	Nickname string
	// 创建时间，毫秒数
	Ctime int64
	// 更新时间，毫秒数
	Utime int64
}
