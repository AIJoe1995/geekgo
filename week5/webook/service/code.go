package service

import (
	"context"
	"fmt"
	"geekgo/week5/webook/repository"
	"geekgo/week5/webook/service/sms"
	"math/rand"
)

var codeTplId = "1877556"

var (
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
)

type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

type codeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func (svc codeService) Send(ctx context.Context, biz string, phone string) error {

	// 生成验证码
	// 验证码存储起来 svc.repo.Store() 从service层调用repository层
	// 使用短信服务发送验证码
	code := svc.generateCode()
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	err = svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
	if err != nil {
		err = fmt.Errorf("发送短信出现异常 %w", err)
	}
	return err

}

func (svc codeService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {

	// 从存储中取出验证码 和前端传入的验证码进行对比 service调用repository
	return svc.repo.Verify(ctx, biz, phone, inputCode)

}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (svc *codeService) generateCode() string {
	// 六位数，num 在 0, 999999 之间，包含 0 和 999999
	num := rand.Intn(1000000)
	// 不够六位的，加上前导 0
	// 000001
	return fmt.Sprintf("%06d", num)
}
