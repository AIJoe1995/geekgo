package main

import (
	"geekgo/week4/webook/internal/repository"
	"geekgo/week4/webook/internal/repository/cache"
	"geekgo/week4/webook/internal/repository/dao"
	"geekgo/week4/webook/internal/service"
	"geekgo/week4/webook/internal/service/sms"
	"geekgo/week4/webook/internal/service/sms/memory"
	"geekgo/week4/webook/internal/web"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 实现短信验证码登录 验证码存放在缓存中 过期时间1min

// 登录需要注册 /users/login_sms 路由 填写手机号 发送验证码 需要验证验证码
// 保持登录态 需要jwt token 登录态 通过gin middleware刷新维护
// 注册用户需要操作数据库 gorm
// wire 依赖注入

// 目录结构 web -> service -> repository -> dao

func initWebServer() *gin.Engine {
	server := gin.Default()
	return server
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:1234@tcp(localhost:3306)/webook?charset=utf8&parseTime=true"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initSmsService() sms.Service {
	return memory.NewService()
}

func main() {
	server := initWebServer()
	db := initDB()
	smsSvc := initSmsService()

	codeCache := cache.NewCodeCache()
	codeRepo := repository.NewCodeRepository(codeCache)
	codeSvc := service.NewCodeService(codeRepo, smsSvc)
	// UserHandler 传入db
	dao := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(dao)
	svc := service.NewUserService(repo)
	uh := web.NewUserHandler(db, svc, codeSvc)
	uh.RegisterRoutes(server)

	server.Run(":8080")

}
