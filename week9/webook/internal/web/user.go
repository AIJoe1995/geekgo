package web

import (
	"errors"
	"fmt"
	"geekgo/week9/webook/internal/domain"
	"geekgo/week9/webook/internal/service"
	ijwt "geekgo/week9/webook/internal/web/jwt"
	"geekgo/week9/webook/pkgs/ginx"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

type UserHandler struct {
	svc              service.UserService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	ijwt.Handler
}

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	//passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)`
)

func NewUserHandler(svc service.UserService, handler ijwt.Handler) *UserHandler {

	return &UserHandler{
		svc:              svc,
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		Handler:          handler,
	}
}

func (uh *UserHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/users")
	//g.POST("/signup", uh.SignUp)
	//g.POST("/login", uh.Login)
	//g.POST("/edit", uh.Edit)
	//g.GET("/profile", uh.Profile)
	g.POST("/refresh_token", uh.RefreshToken)

	g.POST("/signup", ginx.WrapReq[SignUpReq](uh.SignUpV1))
	g.POST("/login", ginx.WrapReq[LoginReq](uh.LoginV1))
	g.POST("/edit", ginx.WrapReq[ProfileReq](uh.EditV1))
	g.GET("/profile", ginx.WrapReq[struct{}](uh.ProfileV1))
	//g.POST("/refresh_token", ginx.WrapReq[struct{}](uh.RefreshTokenV1))

}
func (uh *UserHandler) SignUpV1(ctx *gin.Context, req SignUpReq) (ginx.Result, error) {

	if req.Password != req.ConfirmPassword {

		return ginx.Result{
			Code: 4,
			Msg:  "两次密码不一致",
		}, errors.New("两次密码不一致")
	}
	// 校验邮箱密码格式
	ok, err := uh.emailRegexExp.MatchString(req.Email)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, errors.New("系统错误")
	}
	if !ok {
		return ginx.Result{
			Code: 4,
			Msg:  "邮箱格式不对",
		}, errors.New("邮箱格式不对")
	}
	ok, err = uh.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	if !ok {
		return ginx.Result{
			Code: 4,
			Msg:  "密码格式不对",
		}, errors.New("密码格式不对")
	}

	err = uh.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	// 用户可能已经存在了
	if err == service.ErrUserDuplicate {
		return ginx.Result{
			Code: 4,
			Msg:  "邮箱已注册",
		}, err
	}
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err

	}
	return ginx.Result{
		Msg: "注册成功",
	}, nil
}

func (uh *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次密码不一致")
	}
	// 校验邮箱密码格式
	ok, err := uh.emailRegexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "邮箱格式不对")
		return
	}
	ok, err = uh.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码格式不对")
		return
	}

	err = uh.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	// 用户可能已经存在了
	if err == service.ErrUserDuplicate {
		ctx.String(http.StatusOK, "邮箱已注册")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "注册成功")
}

func (uh *UserHandler) LoginV1(ctx *gin.Context, req LoginReq) (ginx.Result, error) {

	// 检验输入格式

	u, err := uh.svc.Login(ctx, req.Email, req.Password)

	if err == service.ErrInvalidUserOrPassword {
		return ginx.Result{
			Code: 4,
			Msg:  "邮箱或密码错误",
		}, err

	}
	// 设置jwt token
	err = uh.Handler.SetLoginToken(ctx, u.Id)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "SetLoginTokenError",
		}, err
	}
	return ginx.Result{
		Msg: "登录成功",
	}, nil
}

func (uh *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 检验输入格式

	u, err := uh.svc.Login(ctx, req.Email, req.Password)

	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "邮箱或密码错误")
		return
	}
	// 设置jwt token
	err = uh.Handler.SetLoginToken(ctx, u.Id)
	if err != nil {
		return
	}
	ctx.String(http.StatusOK, "登录成功")

}

func (uh *UserHandler) EditV1(ctx *gin.Context, req ProfileReq) (ginx.Result, error) {

	// 要从登录信息头里面拿到tokenStr 然后解出UserClaim Uid
	// 中间件登录鉴权时 把UserClaims放在claims里
	uc, ok := ctx.Get("claims")
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
	userclaims, ok := uc.(*ijwt.UserClaim)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
	}
	uid := userclaims.Uid
	fmt.Println(uid)
	// 根据uid去更新数据库中对应的profile数据
	return ginx.Result{}, nil

}

func (uh *UserHandler) Edit(ctx *gin.Context) {
	type ProfileReq struct {
		NickName string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}
	var req ProfileReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 校验输入

	// 要从登录信息头里面拿到tokenStr 然后解出UserClaim Uid
	// 中间件登录鉴权时 把UserClaims放在claims里
	uc, ok := ctx.Get("claims")
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
	userclaims, ok := uc.(*ijwt.UserClaim)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
	}
	uid := userclaims.Uid
	fmt.Println(uid)
	// 根据uid去更新数据库中对应的profile数据

}

func (uh *UserHandler) ProfileV1(ctx *gin.Context, req struct{}) (ginx.Result, error) {
	uc, ok := ctx.Get("claims")
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
	userclaims, ok := uc.(*ijwt.UserClaim)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
	}
	uid := userclaims.Uid
	fmt.Println(uid)
	// 根据uid去数据库寻找profile记录 返回

	return ginx.Result{}, nil
}

func (uh *UserHandler) Profile(ctx *gin.Context) {
	uc, ok := ctx.Get("claims")
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
	userclaims, ok := uc.(*ijwt.UserClaim)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
	}
	uid := userclaims.Uid
	fmt.Println(uid)
	// 根据uid去数据库寻找profile记录 返回

}

func (uh *UserHandler) RefreshToken(ctx *gin.Context) {
	// 前端调用refresh_token路由传过来的应该是refresh_token
	refreshToken := uh.ExtractTokenStr(ctx)
	var rc ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(refreshToken, &rc, func(token *jwt.Token) (interface{}, error) {
		return ijwt.RtKey, nil
	})
	if err != nil || !token.Valid {
		//zap.L().Error("系统异常", zap.Error(err))
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = uh.CheckSession(ctx, rc.Ssid)
	if err != nil {
		//// 信息量不足
		//zap.L().Error("系统异常", zap.Error(err))
		//// 要么 redis 有问题，要么已经退出登录
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	// 搞个新的 access_token
	err = uh.SetJWTToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		//zap.L().Error("系统异常", zap.Error(err))
		//// 正常来说，msg 的部分就应该包含足够的定位信息
		//zap.L().Error("ijoihpidf 设置 JWT token 出现异常",
		//	zap.Error(err),
		//	zap.String("method", "UserHandler:RefreshToken"))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "刷新成功",
	})
}
