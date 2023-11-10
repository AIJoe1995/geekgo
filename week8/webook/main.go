package main

import (
	"geekgo/week8/webook/internal/repository"
	"geekgo/week8/webook/internal/repository/article"
	"geekgo/week8/webook/internal/repository/cache"
	"geekgo/week8/webook/internal/repository/dao"
	article2 "geekgo/week8/webook/internal/repository/dao/article"
	"geekgo/week8/webook/internal/service"
	"geekgo/week8/webook/internal/web"
	ijwt "geekgo/week8/webook/internal/web/jwt"
	"geekgo/week8/webook/internal/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	dao.InitTable(db)
	return db
}

func main() {
	server := gin.Default()

	db := initDB()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	jwtHdl := ijwt.NewRedisJWTHandler(client)
	mdl := middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
		IgnorePath("/users/signup").
		IgnorePath("/users/refresh_token").
		IgnorePath("/users/login_sms/code/send").
		IgnorePath("/users/login_sms").
		IgnorePath("/oauth2/wechat/authurl").
		IgnorePath("/oauth2/wechat/callback").
		IgnorePath("/users/login").
		IgnorePath("/test/metric").
		Build()

	server.Use(mdl)

	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewRedisUserCache(client)
	userRepo := repository.NewUserRepository(userDAO, userCache)
	userSvc := service.NewUserService(userRepo)
	userHdl := web.NewUserHandler(userSvc, jwtHdl)
	userHdl.RegiterRoutes(server)

	articleDAO := article2.NewArticleDAO(db)
	articleRepo := article.NewArticleRepository(articleDAO)
	articleSvc := service.NewArticleService(articleRepo)

	interDAO := dao.NewGORMInteractiveDAO(db)
	//interCache := cache.NewRedisInteractiveCache(client)
	interCache := cache.NewInteractiveCacheV1(client)
	interRepo := repository.NewCachedReadCntRepository(interCache, interDAO)
	interSvc := service.NewInteractiveService(interRepo)
	artHdl := web.NewArticleHandler(articleSvc, interSvc)
	artHdl.RegisterRoutes(server)

	//server.Use()

	server.Run(":8080")
}
