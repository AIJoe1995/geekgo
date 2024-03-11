package grpc

import (
	"context"
)

type Server struct {
	UnimplementedUserServiceServer
	Name string
}

func (s *Server) GetById(ctx context.Context, request *GetByIdRequest) (*GetByIdResponse, error) {
	return &GetByIdResponse{
		User: &User{
			Id: request.Id,
		},
	}, nil
}
