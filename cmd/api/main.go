package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"library/internal/config"
	"library/internal/handlers"
	"library/internal/repositories"
	"library/internal/services"
)

func main() {
	// 1. Initialize structured logging
	logger := log.New(os.Stdout, "[LIBRARY-API] ", log.LstdFlags|log.Lshortfile)

	// 2. Load system environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Failed to load environment configuration: %v", err)
	}
	logger.Println("Configuration variables parsed successfully.")

	// 3. Initialize production-tuned database pool
	db, err := config.InitDB(cfg.DBConn)
	if err != nil {
		logger.Fatalf("Failed to initialize PostgreSQL connection pool: %v", err)
	}
	defer func() {
		logger.Println("Closing database connection pool...")
		if err := db.Close(); err != nil {
			logger.Printf("Error closing database connection pool: %v", err)
		}
	}()
	logger.Println("Database connection pool established and verified via ping.")

	// 4. Dependency Injection (Clean Architecture Composition Root)
	bookRepo := repositories.NewBookRepository()
	memberRepo := repositories.NewMemberRepository()

	bookService := services.NewBookService(db, bookRepo, memberRepo)
	memberService := services.NewMemberService(db, memberRepo)

	bookHandler := handlers.NewBookHandler(bookService)
	memberHandler := handlers.NewMemberHandler(memberService)

	// 5. Standard library routing registration using ServeMux
	mux := http.NewServeMux()

	// Register the clean, self-multiplexing handlers directly to path trees
	mux.Handle("/books", bookHandler)
	mux.Handle("/books/borrow", bookHandler)
	mux.Handle("/books/return", bookHandler)
	mux.Handle("/members", memberHandler)

	// Global inline middleware for logging API traffic cleanly
	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Printf("%s %s handled in %s", r.Method, r.URL.Path, time.Since(start))
		})
	}

	// 6. Server setup configuration
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.AppPort),
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 7. Graceful Shutdown orchestration
	// Create a channel to catch OS termination signals (SIGINT, SIGTERM)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Spin up the server engine inside a non-blocking background goroutine
	go func() {
		logger.Printf("HTTP Server booting on port %d...", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Critical error during server runtime execution: %v", err)
		}
	}()

	// Block main process flow right here until a shutdown signal drops in
	sig := <-shutdownChan
	logger.Printf("Received termination signal (%s). Initiating graceful shutdown sequence...", sig)

	// Grant the server an operational buffer window of 5 seconds to drain active requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown violently before draining connections: %v", err)
	}

	logger.Println("Server exited cleanly. System offline.")
}
