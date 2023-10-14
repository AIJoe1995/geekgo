package main

import (
	"github.com/gin-gonic/gin"
	"week6/webook/service"
	"week6/webook/service/sms/memory"
	"week6/webook/web"
)

// 实现发送短信验证码的路由 实现短信服务 支持限流和异步发送
// 需要将被限流和发送失败的短信转存到数据库 进行异步发送

// 这里最小demo并不需要userhandler 只需要注册发送短信的路由就可

func main() {
	server := gin.Default()
	smsSvc := memory.NewService()
	codeSvc := service.NewCodeService(smsSvc)
	sms := web.NewSMS(codeSvc)
	server.POST("/send_sms", sms.SendSMSCode)
	server.Run(":8080")
}
