package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-url-shortener/internal/config"
	"go-url-shortener/internal/handler"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/service"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	// Connect PostgreSQL
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Postgres connection failure: %v", err)
	}
	defer db.Close()

	// Construct Database Table Schema directly if missing
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls (
		id BIGSERIAL PRIMARY KEY,
		long_url TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);`)
	if err != nil {
		log.Fatalf("Database migrations failed: %v", err)
	}

	// Connect Redis
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	defer rdb.Close()

	// Construct Core Structural Layers (Dependency Injection)
	repo := repository.NewURLRepository(db, rdb)
	svc := service.NewURLService(repo)
	httpMux := handler.NewHTTPHandler(svc)

	srv := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: httpMux,
	}

	// Graceful Execution Thread Execution Control Setup
	go func() {
		log.Printf("Server online operating at port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen execution break down: %s\n", err)
		}
	}()

	// Block main execution thread state awaiting termination alerts
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Intercepted kill signal. Halting server pools gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced forced to shutdown prematurely: %v", err)
	}

	log.Println("Server exited cleanly.")
}
