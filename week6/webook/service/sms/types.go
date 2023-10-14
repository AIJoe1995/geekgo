package sms

import (
	"context"
)

// types.go里面提供Service接口 在sms下层目录里提供各种不同实现
type Service interface {
	Send(ctx context.Context, tpl string, args []string, phone []string) error
}
