package service

import (
	"context"
	"fmt"
	"math/rand"
	"week6/webook/service/sms"
)

const codeTplId = "1"

type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
}

type codeService struct {
	smsSvc sms.Service
}

func NewCodeService(smsSvc sms.Service) CodeService {
	return &codeService{
		smsSvc: smsSvc,
	}
}

func (c *codeService) Send(ctx context.Context, biz, phone string) error {
	// code service 需要调用sms service的发送方法
	code := c.generateCode()
	return c.smsSvc.Send(ctx, codeTplId, []string{code}, []string{phone})

}

func (c *codeService) generateCode() string {
	num := rand.Intn(999999)
	return fmt.Sprintf("%06d", num)
}
