package main

import (
	"geekgo/week5/webook/repository"
	"geekgo/week5/webook/repository/cache"
	"geekgo/week5/webook/repository/dao"
	"geekgo/week5/webook/service"
	"geekgo/week5/webook/service/sms/memory"
	"geekgo/week5/webook/web"
	ijwt "geekgo/week5/webook/web/jwt"
	"geekgo/week5/webook/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:1234@tcp(localhost:3306)/webook?charset=utf8&parseTime=true"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}

func initWebServer(userHdl *web.UserHandler, jwtHdl ijwt.Handler) *gin.Engine {
	server := gin.Default()

	// 使用jwt 需要放在gin middleware
	mdls := initMiddlewares(jwtHdl)
	server.Use(mdls...)

	userHdl.RegisterRoutes(server)
	return server
}

func initMiddlewares(jwtHdl ijwt.Handler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").
			Build(),
	}
}

func initRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	return redisClient
}

func main() {
	// 启动服务端程序
	db := initDB()
	cmdable := initRedis()
	userDAO := dao.NewUserDAO(db)
	userRepo := repository.NewUserRepository(userDAO)
	userSvc := service.NewUserService(userRepo)

	cache := cache.NewCodeCache(cmdable)
	codeRepo := repository.NewCodeRepository(cache)
	smsSvc := memory.NewService()
	codeSvc := service.NewCodeService(codeRepo, smsSvc)

	jwtHdl := ijwt.NewRedisJWTHandler(cmdable)

	userHdl := web.NewUserHandler(codeSvc, userSvc, jwtHdl)
	server := initWebServer(userHdl, jwtHdl)
	server.Run(":8080")
}
