package startup

import (
	"week6/webook/repository"
	"week6/webook/repository/dao"
	"week6/webook/service/sms"
	"week6/webook/service/sms/async"
)

func InitAsyncSmsService(svc sms.Service) async.AsyncSMSService {
	gormDB := InitTestDB()
	asyncSmsDAO := dao.NewGORMAsyncSmsDAO(gormDB)
	asyncSmsRepository := repository.NewSMSRepository(asyncSmsDAO)

	asyncService := async.NewAsyncSMSService(svc, asyncSmsRepository)
	return asyncService
}
