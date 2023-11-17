package ioc

import (
	"geekgo/week9/webook/internal/web"
	ijwt "geekgo/week9/webook/internal/web/jwt"
	"geekgo/week9/webook/internal/web/middleware"
	"geekgo/week9/webook/pkgs/ginx/metrics"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler, artHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	artHdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(cmd redis.Cmdable, jwtHdl ijwt.Handler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgnorePath("/users/signup").
			IgnorePath("/users/refresh_token").
			IgnorePath("/users/login_sms/code/send").
			IgnorePath("/users/login_sms").
			IgnorePath("/oauth2/wechat/authurl").
			IgnorePath("/oauth2/wechat/callback").
			IgnorePath("/users/login").
			IgnorePath("/test/metric").
			Build(),
		metrics.NewMiddlewareBuilder("week9", "webook", "ginx_http", "ginx_metrics", "1").Build(),
	}
}
