package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
)

type UserHandler struct {
	db *gorm.DB
}

func main() {

	//db := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"), &gorm.Config{})
	db, _ := gorm.Open(mysql.Open("root:root@tcp(webook-live-mysql:3308)/webook"), &gorm.Config{})

	fmt.Printf("%p", db)

	server := gin.Default()

	redisClient := redis.NewClient(&redis.Options{
		//Addr: "localhost:6379",
		Addr: "webook-live-redis:6380",
	})

	fmt.Printf("%p", redisClient)

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "你好你来了")
	})
	server.Run(":8081")
}
