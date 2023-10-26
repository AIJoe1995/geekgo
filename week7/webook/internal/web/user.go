package web

import (
	"geekgo/week7/webook/internal/domain"
	"geekgo/week7/webook/internal/service"
	"geekgo/week7/webook/pkg/ginx"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

type UserHandler struct {
	svc         *service.UserService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	return &UserHandler{
		svc:         svc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
	}
}

func (uh *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	//ug.POST("/signup", uh.SignUp)
	//ug.POST("/login", uh.Login)
	ug.POST("/signup", ginx.WrapBody[SignUpReq](uh.SignUpWrap))
	ug.POST("/login", ginx.WrapBody[LoginReq](uh.LoginWrap))
}

func (uh *UserHandler) SignUpWrap(ctx *gin.Context, req SignUpReq) (Result, error) {

	// 检验两次密码
	if req.Password != req.ConfirmPassword {
		//ctx.JSON(http.StatusOK, )
		return Result{
			Code: 4,
			Msg:  "两次密码不一致",
		}, nil
	}

	// 做格式校验
	if !uh.checkEmailFormat(req.Email) {
		//ctx.JSON(http.StatusOK, )
		return Result{
			Code: 4,
			Msg:  "邮箱格式不对",
		}, nil
	}
	if !uh.checkPasswordFormat(req.Password) {
		//ctx.JSON(http.StatusOK, )
		return Result{
			Code: 4,
			Msg:  "密码格式不对",
		}, nil
	}

	// signup 需要注册用户 向数据库插入数据
	// 业务逻辑 如果是邮箱已经注册的用户 返回邮箱冲突
	// 密码需要加密，密码在那一层来加密？

	err := uh.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	// 检查返回的err 是不是邮箱冲突 如果是返回特定msg
	if err == service.ErrUserDuplicate {
		//ctx.JSON(http.StatusOK, )
		return Result{
			Code: 4,
			Msg:  "邮箱冲突",
		}, err
	}
	if err != nil {
		//ctx.JSON(http.StatusOK, )
		return Result{
			Code: 4,
			Msg:  "系统错误",
		}, err
	}

	// 对于返回的user 需要保持登录态 拿到user id 存到jwt token里 对于注册成功的 redirect到登录页面 维持登录态

	//最后返回
	//ctx.JSON(http.StatusOK, )
	return Result{
		Msg: "注册成功",
	}, nil
}

func (uh *UserHandler) LoginWrap(ctx *gin.Context, req LoginReq) (Result, error) {

	// 需要返回Email Password 对应的user
	user, err := uh.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrUserNotFound {
		//ctx.JSON(http.StatusOK, )
		return Result{
			Code: 5,
			Msg:  "用户名或密码错误",
		}, err
	}

	if err != nil {
		//ctx.JSON(http.StatusOK, )
		return Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	userclaims := UserClaims{
		UserID: user.Id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES512, userclaims)
	tokenStr, err := token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "系统错误",
		})
	}
	ctx.Header("x-jwt-token", tokenStr)

	//ctx.JSON(http.StatusOK, )
	return Result{
		Msg: "登陆成功"}, nil
}

func (uh *UserHandler) checkEmailFormat(email string) bool {
	ok, _ := uh.emailExp.MatchString(email)
	return ok
}

func (uh *UserHandler) checkPasswordFormat(password string) bool {
	ok, _ := uh.passwordExp.MatchString(password)
	return ok
}

func (uh *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}
	// 需要从ctx里拿到请求的数据
	var req SignUpReq
	err := ctx.Bind(&req)
	if err != nil {
		return
	}
	// 检验两次密码
	if req.Password != req.ConfirmPassword {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "两次密码不一致",
		})
		return
	}

	// 做格式校验
	if !uh.checkEmailFormat(req.Email) {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "邮箱格式不对",
		})
		return
	}
	if !uh.checkPasswordFormat(req.Password) {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "密码格式不对",
		})
		return
	}

	// signup 需要注册用户 向数据库插入数据
	// 业务逻辑 如果是邮箱已经注册的用户 返回邮箱冲突
	// 密码需要加密，密码在那一层来加密？

	err = uh.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	// 检查返回的err 是不是邮箱冲突 如果是返回特定msg
	if err == service.ErrUserDuplicate {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "邮箱冲突",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "系统错误",
		})
		return
	}

	// 对于返回的user 需要保持登录态 拿到user id 存到jwt token里 对于注册成功的 redirect到登录页面 维持登录态

	//最后返回
	ctx.JSON(http.StatusOK, Result{
		Msg: "注册成功",
	})
}

func (uh *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	// 需要返回Email Password 对应的user
	user, err := uh.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrUserNotFound {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "用户名或密码错误",
		})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	userclaims := UserClaims{
		UserID: user.Id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES512, userclaims)
	tokenStr, err := token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "系统错误",
		})
	}
	ctx.Header("x-jwt-token", tokenStr)

	ctx.JSON(http.StatusOK, Result{
		Msg: "登陆成功"})
	return
}

type UserClaims struct {
	jwt.RegisteredClaims
	UserID int64
}
