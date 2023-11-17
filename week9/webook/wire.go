//go:build wireinject

package main

import (
	"geekgo/week9/webook/internal/events/article"
	"geekgo/week9/webook/internal/repository"
	"geekgo/week9/webook/internal/repository/cache"
	"geekgo/week9/webook/internal/repository/dao"
	"geekgo/week9/webook/internal/service"
	"geekgo/week9/webook/internal/web"
	ijwt "geekgo/week9/webook/internal/web/jwt"
	"geekgo/week9/webook/ioc"
	"github.com/google/wire"
)

func InitServer() *App {
	wire.Build(
		ioc.InitDB,
		ioc.InitRedis,
		dao.NewGORMUserDAO,
		cache.NewRedisUserCache,
		repository.NewUserRepository,
		service.NewUserService,
		web.NewUserHandler,
		ijwt.NewRedisJWT,
		ioc.InitWebServer,

		dao.NewGORMArticleDAO,
		repository.NewArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler,

		ioc.InitMiddlewares,

		ioc.InitKafka,
		ioc.NewConsumers,
		ioc.NewSyncProducer,

		article.NewKafkaProducer,
		article.NewInteractiveReadEventConsumer,

		dao.NewGORMInteractiveDAO,
		cache.NewRedisInteractiveCache,
		repository.NewCachedReadCntRepository,
		service.NewInteractiveService,

		ioc.InitLogger,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
