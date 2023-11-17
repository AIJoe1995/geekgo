package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

var (
	AtKey = []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0")
	RtKey = []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvfx")
)

type RedisJWT struct {
	cmd redis.Cmdable
}

func NewRedisJWT(cmd redis.Cmdable) Handler {
	return &RedisJWT{
		cmd: cmd,
	}
}

func (r *RedisJWT) ExtractTokenStr(ctx *gin.Context) string {
	auth := ctx.GetHeader("Authorization")

	slice := strings.Split(auth, " ")
	if len(slice) != 2 {
		return ""
	}
	return slice[1]
}
func (r *RedisJWT) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	claims := ctx.MustGet("claims").(*UserClaim)
	// 把失效的sessionid 存到redis 因为jwt本身是无状态的， 如果之后认为这个jwttoken已经失效了 需要额外的地方记录 这里选择在redis里记录ssid
	return r.cmd.Set(ctx, fmt.Sprintf("users:ssid:%s", claims.Ssid), "", time.Hour*24*7).Err()

}

// 引入ssid的目的在于 退出登录时 需要设置长token无效或者需要把长短token都设置为"" 都设置为无效，
// 这种情况登录校验 就需要校验长token是否还有效，那每次请求都需要携带长token
// 能不能用一个东西标识这一次登录？这就是ssid, 最后只检验ssid是不是还有效。
// login之后需要设置 长短token 前端在401时 发送长token回来 验证长token ok后，重新设置短token
func (r *RedisJWT) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := r.SetJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	err = r.setRefreshToken(ctx, uid, ssid)
	return err
}

func (r *RedisJWT) CheckSession(ctx *gin.Context, ssid string) error {
	val, err := r.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	switch err {
	case redis.Nil:
		return nil
	case nil:
		if val == 0 {
			return nil
		}
		return errors.New("session已经失效")
	default:
		return err
	}
}

func (r *RedisJWT) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	// 设置jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaim{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
	})
	tokenStr, err := token.SignedString(AtKey)
	if err != nil {
		return err
	}
	// tokenStr 放到header里
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (r *RedisJWT) setRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	refreshClaims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid:  uid,
		Ssid: ssid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshClaims)
	tokenStr, err := token.SignedString(RtKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}
