package main

import (
	"log"

	pb "llyb-backend/proto"
	"trpc.group/trpc-go/trpc-go"
	_ "trpc.group/trpc-go/trpc-go/http"
)

func main() {
	server := trpc.NewServer()
	service := server.Service(pb.AdminServiceServer_ServiceDesc.ServiceName)
	pb.RegisterAdminService(service, &AdminService{})
	if err := server.Serve(); err != nil {
		log.Fatalf("trpc server exit: %v", err)
	}
}
