package ginx

import (
	"geekgo/week7/webook/internal/web"
	"geekgo/week7/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

var l logger.LoggerV1

func WrapBody[Req any](fn func(ctx *gin.Context, req Req) (web.Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			return
		}
		res, err := fn(ctx, req)
		if err != nil {
			l.Error("处理业务逻辑出错", logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err))

		}
		ctx.JSON(http.StatusOK, res)
	}
}
