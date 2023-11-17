package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	ijwt "geekgo/week9/webook/internal/web/jwt"
	"net/http"
)

// 每一次请求都要进行登录鉴权 除了一些特定的路由

type LoginJWTMiddlewareBuilder struct {
	paths []string
	ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(jwtHdl ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: jwtHdl,
	}
}

func (j *LoginJWTMiddlewareBuilder) IgnorePath(path string) *LoginJWTMiddlewareBuilder {
	j.paths = append(j.paths, path)
	return j
}

func (j *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	// 返回的是一个函数
	return func(ctx *gin.Context) {
		for _, path := range j.paths {
			if path == ctx.Request.URL.Path {
				return
			}
		}

		tokenStr := j.ExtractTokenStr(ctx)
		var uc = ijwt.UserClaim{}
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return ijwt.AtKey, nil
		})
		if err != nil || !token.Valid || token == nil || uc.Uid == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 检查用户是否退出登录了
		err = j.CheckSession(ctx, uc.Ssid)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// jwt AccessKey RefreshKey 自动续期
		//expireTime, err := uc.GetExpirationTime()
		//if err != nil {
		//	// 拿不到过期时间
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}
		//if expireTime.Before(time.Now()) {
		//	// 已经过期
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}
		//
		//// 刷新过期时间
		//now := time.Now()
		//if uc.ExpiresAt.Sub(now) < time.Second*50 {
		//	uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
		//	tokenStr, err = token.SignedString(web.JWTKey)
		//	if err != nil {
		//		// 记录日志
		//	}
		//	ctx.Header("x-jwt-token", tokenStr)
		//}

		ctx.Set("claims", uc)

	}

}
