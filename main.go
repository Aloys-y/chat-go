package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Aloys-y/chat-go/auth"
	"github.com/Aloys-y/chat-go/config"
	"github.com/Aloys-y/chat-go/db"
	"github.com/Aloys-y/chat-go/services"
	"github.com/Aloys-y/chat-go/signaling"

	pb "github.com/Aloys-y/chat-go/proto"

	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	// 创建认证拦截器
	authInterceptor := auth.NewAuthInterceptor()

	// Set up gRPC server with interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.UnaryInterceptor),
	)

	// Register services
	pb.RegisterUserServiceServer(grpcServer, &services.UserServiceImpl{})
	pb.RegisterRoomServiceServer(grpcServer, &services.RoomServiceImpl{})

	// Start gRPC server
	grpcAddr := fmt.Sprintf(":%d", config.AppConfig.Server.GRPCPort)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
	log.Printf("gRPC server listening on %s", grpcAddr)

	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start WebSocket server
	go signaling.StartWSServer(config.AppConfig.Server.WSPort)

	// Wait for interrupt signal to gracefully shut down the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	// Stop gRPC server
	grpcServer.GracefulStop()

	// Stop WebSocket server (will be handled automatically when main exits)

	log.Println("Servers exited gracefully")
}
