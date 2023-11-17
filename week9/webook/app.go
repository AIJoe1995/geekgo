package main

import (
	events "geekgo/week9/webook/internal/events/article"
	"github.com/gin-gonic/gin"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
}
