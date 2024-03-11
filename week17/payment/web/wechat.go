package web

import (
	"geekgo/week17/payment/service/wechat"
	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"net/http"
)

type WechatHandler struct {
	handler *notify.Handler
	// 处理回调路由 调用nativeSvc 如把订单状态等更新到数据库
	nativeSvc *wechat.NativePaymentService
}

func NewWechatHandler(handler *notify.Handler, nativeSvc *wechat.NativePaymentService) *WechatHandler {
	return &WechatHandler{
		handler:   handler,
		nativeSvc: nativeSvc,
	}
}

func (h *WechatHandler) RegisterRoutes(server *gin.Engine) {
	server.Any("/pay/callback", h.HandleNative)
}

func (h *WechatHandler) HandleNative(ctx *gin.Context) {
	transaction := &payments.Transaction{}
	_, err := h.handler.ParseNotifyRequest(ctx, ctx.Request, transaction)
	if err != nil {
		ctx.JSON(http.StatusOK, err.Error())
	}
	err = h.nativeSvc.HandleCallback(ctx, transaction)
	if err != nil {
		ctx.JSON(http.StatusOK, err.Error())
	}
	ctx.JSON(http.StatusOK, "ok")
}
