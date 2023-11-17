package ioc

import (
	"geekgo/week9/webook/internal/repository/dao"
	"geekgo/week9/webook/pkgs/gormx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook?parseTime=True&&charset=utf8"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)

	db.Use(gormx.NewPlugin())

	if err != nil {
		panic(err)
	}
	return db
}
