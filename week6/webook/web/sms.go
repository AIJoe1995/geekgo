package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"week6/webook/service"
)

const biz = "login"

type SMS struct {
	codeSvc service.CodeService
}

func NewSMS(codeSvc service.CodeService) *SMS {
	return &SMS{
		codeSvc: codeSvc,
	}
}

func (sms *SMS) SendSMSCode(ctx *gin.Context) {
	// 调用service层的codeservice 来发送短信
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	err := sms.codeSvc.Send(ctx, biz, req.Phone)
	if err != nil {
		ctx.String(http.StatusOK, "发送短信失败")
	}
	ctx.String(http.StatusOK, "发送短信成功")

}
