package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"google.golang.org/grpc"

	"github.com/kuberoute/orchestrator/internal/api"
	"github.com/kuberoute/orchestrator/internal/auth"
	"github.com/kuberoute/orchestrator/internal/clusters"
	"github.com/kuberoute/orchestrator/internal/db"
	"github.com/kuberoute/orchestrator/internal/decisions"
	"github.com/kuberoute/orchestrator/internal/optimize"
	pb "github.com/kuberoute/orchestrator/proto"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database
	dbPool, err := pgxpool.New(ctx, getEnv("DATABASE_URL", "postgres://kuberoute:kuberoute_dev@localhost:5432/kuberoute?sslmode=disable"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	if err := db.Migrate(ctx, dbPool); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_URL", "localhost:6379"),
	})
	defer rdb.Close()

	// NATS
	nc, err := nats.Connect(getEnv("NATS_URL", "nats://localhost:4222"))
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	js, _ := nc.JetStream()

	// Auth
	jwtSecret, err := auth.ResolveJWTSecret(getEnv("KUBEROUTE_ENV", "local"), os.Getenv("JWT_SECRET"))
	if err != nil {
		log.Fatal(err)
	}
	authService := auth.NewService(dbPool, auth.Config{
		JWTSecret:          jwtSecret,
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
	})

	// Repositories
	clusterRepo := clusters.NewRepository(dbPool)
	decisionRepo := decisions.NewRepository(dbPool)

	// Services
	clusterService := clusters.NewService(clusterRepo)
	decisionService := decisions.NewService(decisionRepo, js)

	// gRPC Server
	grpcServer := grpc.NewServer()
	optimizer := optimize.NewClient(getEnv("AI_ENGINE_URL", "http://localhost:8081"))
	agentServer := api.NewAgentServer(clusterService, decisionService, dbPool, optimizer)
	pb.RegisterAgentServiceServer(grpcServer, agentServer)

	grpcAddr := getEnv("GRPC_ADDR", ":50051")
	go func() {
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Fatalf("failed to listen on %s: %v", grpcAddr, err)
		}
		log.Printf("gRPC server listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// REST Router
	router := mux.NewRouter()
	apiHandler := api.NewRESTHandler(authService, clusterService, decisionService)
	apiHandler.RegisterRoutes(router)

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			getEnv("DASHBOARD_URL", "http://localhost:3000"),
			"https://app.kuberoute.ai",
		},
		AllowOriginFunc: func(origin string) bool {
			return origin == "http://localhost:3000" ||
				len(origin) > 16 && origin[:16] == "http://localhost" ||
				origin == "https://app.kuberoute.ai"
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	httpAddr := getEnv("HTTP_ADDR", ":8080")
	server := &http.Server{
		Addr:         httpAddr,
		Handler:      c.Handler(router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("HTTP server listening on %s", httpAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down servers...")
	grpcServer.GracefulStop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
