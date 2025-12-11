package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/chat-go/config"
	"github.com/yourusername/chat-go/db"
	"github.com/yourusername/chat-go/models"
	"github.com/yourusername/chat-go/signaling"
	"github.com/yourusername/chat-go/services"

	pb "github.com/yourusername/chat-go/proto"

	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	if err := db.InitDB(cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	// Auto migrate models
	log.Println("Running database migrations...")
	if err := models.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Set up gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	pb.RegisterUserServiceServer(grpcServer, &services.UserServiceImpl{})
	pb.RegisterRoomServiceServer(grpcServer, &services.RoomServiceImpl{})

	// Start gRPC server
	grpcAddr := fmt.Sprintf(":%d", cfg.Server.GRPCPort)
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
	go signaling.StartWSServer(cfg.Server.WSPort)

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