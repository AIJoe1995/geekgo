package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"geekgo/week7/webook/internal/repository"
	"geekgo/week7/webook/internal/repository/dao"
	"geekgo/week7/webook/internal/service"
	"geekgo/week7/webook/internal/web"
)

func main() {
	server := gin.Default()
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	// 验证登录态的middleware 需要忽略一些路径 如signup login 使用Builder模式
	server.Use()
	dao := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(dao)
	svc := service.NewUserService(repo)
	uh := web.NewUserHandler(svc)
	uh.RegisterRoutes(server)
	server.Run(":8080")
}
