package main

// 使用gorm 需要 go get -u gorm.io/gorm
//go get -u gorm.io/driver/mysql
import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type User struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"'`
	Email    string `gorm:"unique"`
	Password string
	Utime    int64
	Ctime    int64
	Nickname string
	Birthday time.Time
	AboutMe  string
}

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}

func initDB() *gorm.DB {
	// [user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
	db, err := gorm.Open(mysql.Open("root:1234@tcp(localhost:3306)/webook?charset=utf8&parseTime=true"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initServer() *gin.Engine {
	server := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	server.Use(sessions.Sessions("mysession", store))
	return server
}

var (
	emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
)

type UserHandler struct {
	db *gorm.DB
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	// signup 从ctx中获取 对应的数据 存到结构体变量中 然后传递结构体变量 需要在数据库中插入一条记录
	// SignUp 的操作 是从web 到service 到repository 到dao一层层调用的 dao中处理操作
	// 简单的方式是SignUp所有逻辑都在这里处理 之后再进行拆分
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	var req SignUpReq

	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次密码不同")
		return
	}

	// 逻辑是 需要先进行密码和邮箱格式的校验
	emailPattern := regexp.MustCompile(emailRegexPattern, regexp.None)
	isEmail, err := emailPattern.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "邮箱格式不正确")
		return
	}

	passwordPattern := regexp.MustCompile(passwordRegexPattern, regexp.None)
	isPassword, err := passwordPattern.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码格式不正确")
		return
	}

	// 邮箱 密码格式校验完成之后 需要注册账号 既需要在数据库中插入一条数据 插入数据的时候 密码需要加密
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return
	}

	// 需要确认数据库里有没有这个用户
	type User struct {
		Id       int64
		Email    string
		Password string
		Utime    int64
		Ctime    int64
		Nickname string
		Birthday time.Time
		AboutMe  string
	}
	user := User{}
	user.Email = req.Email
	user.Password = string(hash)
	now := time.Now().UnixMilli()
	user.Utime = now
	user.Ctime = now

	// 需要使用db 看起来应该建一个结构体 结构体里面包含db实例 这种结构体作为方法接受者 这样才方便在这个函数里使用db

	result := u.db.Create(&user)
	// user.Id
	// result.Error
	// result.RowsAffected
	if result.Error != nil {
		ctx.String(http.StatusOK, "用户创建失败")
	}
	ctx.String(http.StatusOK, "用户创建成功")

}

func (u *UserHandler) Login(ctx *gin.Context) {

	//  从ctx获取登录页面填的账号密码 和数据库中的账号密码 进行比对 写入session
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	err := ctx.Bind(&req)
	if err != nil {
		return
	}
	type User struct {
		Id       int64
		Email    string
		Password string
		Utime    int64
		Ctime    int64
		Nickname string
		Birthday time.Time
		AboutMe  string
	}
	user := User{}
	result := u.db.First(&user, "email = ?", req.Email)
	if result.Error != nil {
		ctx.String(http.StatusOK, "查不到用户或其他错误")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		ctx.String(http.StatusOK, "用户名或密码错误")
	}

	// 使用session记录登录状态
	sess := sessions.Default(ctx)
	sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		// 60 秒过期
		MaxAge: 3600,
	})
	err = sess.Save()
	if err != nil {
		ctx.String(http.StatusOK, "服务器异常")
		return
	}
	ctx.String(http.StatusOK, "登录成功")

}

func (u *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}
	// Edit需要获取sessionid对应的用户的id
	sess := sessions.Default(ctx)
	id := sess.Get("userId").(int64)
	var req EditReq
	err := ctx.Bind(&req)
	if err != nil {
		ctx.String(http.StatusOK, "传参问题或内部错误")
	}

	// 把id对应的user中req对应的字段更新

	type User struct {
		Id       int64
		Email    string
		Password string
		Utime    int64
		Ctime    int64
		Nickname string
		Birthday time.Time
		AboutMe  string
	}
	user := User{}
	result := u.db.First(&user, "id = ?", id)
	if result.Error != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	user.Nickname = req.Nickname
	user.Birthday = carbon.Parse(req.Birthday).Carbon2Time() // 这里对不能parse的格式没有返回错误信息 这样处理不够好
	if err != nil {
		ctx.String(http.StatusOK, "输入的birthday格式不对，应该是xxxx-xx-xx")
		return
	}
	user.AboutMe = req.AboutMe
	u.db.Save(&user)
	ctx.String(http.StatusOK, "profile保存成功")
}

func (u *UserHandler) Profile(ctx *gin.Context) {

	// 展示Profile FindById 返回哪些信息？
	sess := sessions.Default(ctx)
	id := sess.Get("userId").(int64)
	type User struct {
		Id       int64
		Email    string
		Password string
		Utime    int64
		Ctime    int64
		Nickname string
		Birthday time.Time
		AboutMe  string
	}
	user := User{}
	result := u.db.First(&user, "id = ?", id)
	if result.Error != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	type UserProfile struct {
		Nickname string
		Birthday time.Time
		AboutMe  string
	}
	ctx.JSON(http.StatusOK, UserProfile{
		Nickname: user.Nickname,
		Birthday: user.Birthday,
		AboutMe:  user.AboutMe,
	})

}

func (u *UserHandler) RegisterUserRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.Login)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.Profile)

}

func main() {
	// 需要实现的功能 用户的注册 登录 修改profile
	// 需要注册的路由 路由组 /users  用户路由组下 需要注册 signup login profile edit 路由 通过gin.Engine..GET等注册  需要创建gin Engine
	// 用户注册需要进行password和email的格式校验 需要regex库
	// 用户登录 需要获取session 根据session执行登录校验 （用户登录后需要保持登录态） 需要使用session插件
	// 用户密码等敏感信息存入数据库之前需要进行加密 需要bcrypt库
	// 用户需要 Email Password Id Utime Ctime Birthday NickName Phone Age AboutMe  等字段 需要ORM 通过gorm来访问数据库 需要创建数据库连接
	db := initDB()
	server := initServer()

	u := UserHandler{db: db}
	// 需要在server中注册users路由组 再向组中注册路由
	u.RegisterUserRoutes(server)

	// 注册号路由后 实现对应的SignUp Login Edit Profile方法  // 高阶 应该把这些方法定义在UserHandler这个类型上

	server.Run(":8080")

}
