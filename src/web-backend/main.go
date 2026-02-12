package main

import (
	"context"
	"log"
	"time"

	pb "llyb-backend/proto"
	"llyb-backend/src/chat"
	appinit "llyb-backend/src/init"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/filter"
	_ "trpc.group/trpc-go/trpc-go/http"
	thttp "trpc.group/trpc-go/trpc-go/http"
	"trpc.group/trpc-go/trpc-go/server"
)

func main() {
	db, err := appinit.OpenMySQLFromEnv()
	if err != nil {
		log.Fatalf("mysql connect failed: %v", err)
	}
	{
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		if err := appinit.EnsureAdminAccountTable(ctx, db); err != nil {
			cancel()
			log.Fatalf("mysql ensure schema failed: %v", err)
		}
		cancel()
	}

	// Avoid CORS preflight during dev by letting clients send JSON with a simple
	// Content-Type (text/plain). We still return JSON.
	thttp.SetContentType("text/plain", codec.SerializationTypeJSON)

	corsFilter := func(ctx context.Context, req any, next filter.ServerHandleFunc) (any, error) {
		rw := thttp.Response(ctx)
		if rw != nil {
			// Dev-friendly CORS; use "*" since we don't rely on cookies right now.
			rw.Header().Set("Access-Control-Allow-Origin", "*")
			rw.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
			rw.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		return next(ctx, req)
	}

	s := trpc.NewServer(server.WithFilters([]filter.ServerFilter{corsFilter}))
	// trpc-go codegen exports the service descriptor as AdminServer_ServiceDesc.
	service := s.Service(pb.AdminServer_ServiceDesc.ServiceName)
	if service == nil {
		log.Fatalf("trpc service %q not found; check trpc_go.yaml server.service[].name", pb.AdminServer_ServiceDesc.ServiceName)
	}
	pb.RegisterAdminService(service, &AdminService{db: db})

	// Coexistence on the same port:
	// - Existing endpoints (/admin/login, /admin/register) are HTTP-RPC methods generated from proto.
	// - AI streaming endpoint is a standard HTTP handler registered into the same service.
	//
	// This avoids adding another listener/port and keeps routing in one place.
	thttp.HandleFunc("/ai/chat/stream", chat.StreamHandler)
	thttp.RegisterNoProtocolService(service)

	if err := s.Serve(); err != nil {
		log.Fatalf("trpc server exit: %v", err)
	}
}
