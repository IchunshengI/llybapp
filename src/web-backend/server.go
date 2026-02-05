package main

import (
	"context"
	"database/sql"
	"log"

	"llyb-backend/login"
	pb "llyb-backend/proto"
)

// AdminService keeps all interface handlers in one file for now.
// Embed for forward compatibility with generated interfaces.
type AdminService struct {
	pb.UnimplementedAdmin

	db *sql.DB
}

func (s *AdminService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	ok, msg, err := login.Login(ctx, s.db, req.GetUsername(), req.GetPassword())
	if err != nil {
		log.Printf("login failed: username=%q err=%v", req.GetUsername(), err)
		return &pb.LoginResponse{Ok: false, Message: "系统错误"}, nil
	}
	return &pb.LoginResponse{Ok: ok, Message: msg}, nil
}

func (s *AdminService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	res, err := login.Register(ctx, s.db, req.GetUsername(), req.GetPassword())
	if err != nil {
		log.Printf("register failed: username=%q err=%v", req.GetUsername(), err)
	}
	return &pb.RegisterResponse{
		Code:      res.Code,
		AccountId: res.AccountID,
		Message:   res.Message,
	}, nil
}
