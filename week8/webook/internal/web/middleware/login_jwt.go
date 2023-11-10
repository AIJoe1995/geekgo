package middleware

import (
	ijwt "geekgo/week8/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

// 检验登录态
// 延长jwt过期时间
// 某些请求不需要校验登录态

// 返回gin.HandlerFunc func(ctx *gin.Context)

type LoginJWTMiddlewareBuilder struct {
	paths []string
	ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(jwtHdl ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: jwtHdl,
	}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePath(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		tokenHeader := ctx.GetHeader("Authorization")

		segs := strings.Split(tokenHeader, " ")
		if len(segs) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		claims := &ijwt.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"), nil
		})

		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if token == nil || !token.Valid || claims.Id == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//if claims.UserAgent != ctx.Request.UserAgent() {
		//	// 严重的安全问题
		//	// 你是要监控
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}

		// token
		err = l.CheckSession(ctx, claims.Ssid)
		if err != nil {
			// 要么 redis 有问题，要么已经退出登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//now := time.Now()
		////  每十秒钟刷新一次
		//if claims.ExpiresAt.Sub(now) < time.Second*50 {
		//	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
		//	tokenStr, err = token.SignedString(l.jwtKey)
		//	if err != nil {
		//		log.Println("jwt 续约失败", err)
		//	}
		//	ctx.Header("x-jwt-token", tokenStr)
		//}
		ctx.Set("claims", claims)

	}
}
