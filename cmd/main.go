package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"library/internal/handlers"
	"library/internal/repositories"
	"library/internal/services"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Pure Go PostgreSQL database driver registration
)

func main() {
	// 1. Establish database connection configuration string
	// (Adjust values if your local Postgres instance setup uses different credentials)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, falling back to system environment variables")
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "host=localhost port=5432 user=postgres password=secret dbname=library sslmode=disable"
	}

	log.Println("Connecting to the library inventory database...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Fatal: Failed to open structural database connection pool: %v", err)
	}
	defer db.Close()

	// Configure fundamental sanity rules for connection retention
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Explicitly verify the network connection is functional before running handlers
	if err := db.Ping(); err != nil {
		log.Fatalf("Fatal: Database unreachable via connection string ping check: %v", err)
	}
	log.Println("Database connection successfully established and verified.")

	// 2. Instantiate data repositories
	bookRepo := repositories.NewBookRepository()
	memberRepo := repositories.NewMemberRepository()

	// 3. Inject repositories to bootstrap business workflows
	bookService := services.NewBookService(db, bookRepo, memberRepo)
	memberService := services.NewMemberService(db, memberRepo)

	// 4. Bind services into HTTP server payload multiplexer routers
	bookHandler := handlers.NewBookHandler(bookService)
	memberHandler := handlers.NewMemberHandler(memberService)

	// 5. Explicitly declare routing patterns to avoid route clashes
	mux := http.NewServeMux()

	// Resource REST routers (handles listings, creations, updates, and deletions)
	mux.Handle("/books", bookHandler)
	mux.Handle("/books/", bookHandler)
	mux.Handle("/members", memberHandler)
	mux.Handle("/members/", memberHandler)

	// Custom atomic transaction processing paths
	mux.HandleFunc("/borrow", bookHandler.HandleBorrow)
	mux.HandleFunc("/return", bookHandler.HandleReturn)
	mux.HandleFunc("/loans", bookHandler.HandleListLoans)

	// Global simple middleware layer wrapper for execution request logging
	loggingMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Incoming HTTP Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		mux.ServeHTTP(w, r)
		log.Printf("Completed HTTP Request: %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})

	// 6. Bind listener and run the active network service loop
	port := "8080"
	serverAddr := fmt.Sprintf(":%s", port)
	log.Printf("Library API Management engine active on network address %s", serverAddr)

	if err := http.ListenAndServe(serverAddr, loggingMux); err != nil {
		log.Fatalf("Fatal microservice termination crash: %v", err)
	}
}
