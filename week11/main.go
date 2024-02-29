package week11

import (
	"context"
	"geekgo/week11/ioc"
	"geekgo/week11/repository"
	"geekgo/week11/repository/dao"
	"geekgo/week11/service"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	local := ioc.InitLocalFuncExecutor()
	db, err := gorm.Open(mysql.Open(""), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	d := dao.NewCronJobDAO(db)
	repo := repository.NewCronJobRepository(d)
	svc := service.NewCronJobService(repo)
	schedular := ioc.InitScheduler(local, svc)
	schedular.Schedule(context.Background())
}
