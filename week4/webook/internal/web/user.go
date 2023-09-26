package web

import (
	"fmt"
	"geekgo/week4/webook/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

const biz = "login"

type UserHandler struct {
	db      *gorm.DB
	codeSvc service.CodeService
	svc     service.UserService
}

// codeSvc
func NewUserHandler(db *gorm.DB, svc service.UserService, codeSvc service.CodeService) *UserHandler {
	return &UserHandler{
		db:      db,
		codeSvc: codeSvc,
		svc:     svc,
	}
}

func (uh UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/login_sms", uh.LoginSMS)
	ug.POST("/send_sms_code", uh.SendSMSCode)

}

func (uh UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	// 需要和存下来的code比较 存在phone对应的code且一致且未过期时 能够登录 之后需要转成登录态
	// 如果是新用户 需要插入用户
	// 不是新用户就直接转入登录态
	verify, err := uh.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		if err == service.ErrCodeNotFound {
			ctx.String(http.StatusOK, "请先发送验证码")
		} else if err == service.ErrCodeVerifyTooMany {
			ctx.String(http.StatusOK, "验证太频繁")
		} else {
			ctx.String(http.StatusOK, "未知错误")
		}
		return
	}
	if !verify {
		ctx.String(http.StatusOK, "验证码错误")
		return
	}

	// 验证成功 新用户需要插入数据库 老用户直接登录
	user, err := uh.svc.FindOrCreate(ctx, req.Phone)
	fmt.Printf("%v\n", user)
	ctx.JSON(http.StatusOK, "验证码校验通过")

}

func (uh UserHandler) SendSMSCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		return
	}
	// 检查输入
	if req.Phone == "" {
		ctx.String(http.StatusOK, "输入有误")
		return
	}

	// 需要利用短信服务 smsService 来 发送短信 code需要保存到redis或内存 以便之后verify
	// 一分钟内没有发过验证码 才能发送验证码 不然需要提示验证码发送频繁
	//
	err = uh.codeSvc.Send(ctx, biz, req.Phone)
	if err != nil {
		if err == service.ErrCodeSendTooMany {
			ctx.String(http.StatusOK, "验证码发送太频繁")
		} else {
			ctx.String(http.StatusOK, "发送验证码失败")
		}
		return
	}
	ctx.String(http.StatusOK, "验证码发送成功")

}
