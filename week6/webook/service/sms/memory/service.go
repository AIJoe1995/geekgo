package memory

import (
	"context"
	"fmt"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Send(ctx context.Context, tpl string, args []string, phone []string) error {
	fmt.Println(args)
	return nil
}
