package main

import (
	"context"

	"llyb-backend/login"
	pb "llyb-backend/proto"
)

// AdminService keeps all interface handlers in one file for now.
// Embed for forward compatibility with generated interfaces.
type AdminService struct {
	pb.UnimplementedAdminService
}

func (s *AdminService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	ok, message := login.Handle(req)
	return &pb.LoginResponse{
		Ok:      ok,
		Message: message,
	}, nil
}
