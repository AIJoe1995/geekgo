package web

import (
	"geekgo/week8/webook/internal/domain"
	"geekgo/week8/webook/internal/service"
	ijwt "geekgo/week8/webook/internal/web/jwt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"

	"net/http"
)

type UserHandler struct {
	svc         *service.UserService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
	ijwt.Handler
}

func (h *UserHandler) RegiterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/login", h.Login)
	ug.POST("/signup", h.SignUp)
	ug.GET("/profile", h.Profile)
	ug.GET("/logout", h.Logout)
}

func (h *UserHandler) Logout(ctx *gin.Context) {
	// 清理jwttoken
	err := h.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "退出登录失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "退出登录OK",
	})
}

func (h *UserHandler) Profile(ctx *gin.Context) {
	// 通过middleware的登录校验
	// 从ctx header里 parse Jwt 取出UserId
	type Profile struct {
		Nickname string
	}
	c, _ := ctx.Get("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		// 你可以考虑监控住这里
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	uid := claims.Id
	profile, err := h.svc.Profile(ctx, uid)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Profile{
		Nickname: profile.Nickname,
	})

}

func (h *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 校验
	if req.Email == "" {
		ctx.String(http.StatusOK, "邮箱")
	}

	// 校验邮箱密码格式
	ok, err := h.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "邮箱格式不对")
	}
	ok, err = h.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码格式不对")
	}
	user, err := h.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrUserNotFound {
		ctx.String(http.StatusOK, "邮箱未注册")
		return
	}
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "邮箱或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	// 登陆成功设置登录态
	//uc := UserClaims{
	//	Userid: user.Id,
	//	RegisteredClaims: jwt.RegisteredClaims{
	//		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
	//	},
	//}
	//token := jwt.NewWithClaims(jwt.SigningMethodES512, uc)
	//tokenStr, err := token.SignedString(h.jwtKey)
	//ctx.Header("x-jwt-token", tokenStr)

	if err = h.SetLoginToken(ctx, user.Id); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "登录成功")
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	var req SignUpReq
	// 测试apipost7 header带有Content-type Application/json 这里Bind会返回空字符串 而BindJSON可以返回正常结果
	if err := ctx.Bind(&req); err != nil {
		return
	}
	//fmt.Println(req)
	//// 测试apipost7 header没有Content-type Application/json 这里BindJSON和Bind可以返回正常结果， 但是Bind只能用一次，再次bind就是空字符串。
	//var req SignUpReq
	//ctx.BindJSON(&req)
	//fmt.Println(req)
	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次密码不一致")
	}
	// 校验邮箱密码格式
	ok, err := h.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "邮箱格式不对")
		return
	}
	ok, err = h.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码格式不对")
		return
	}

	err = h.svc.SignUp(ctx, domain.User{
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

func NewUserHandler(svc *service.UserService, jwtHdl ijwt.Handler) *UserHandler {
	const (
		emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		//passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	return &UserHandler{
		svc:         svc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
		Handler:     jwtHdl, // 在UserHandler里组合了ijwt.Handler接口，这里还是要注入具体实现
	}
}
