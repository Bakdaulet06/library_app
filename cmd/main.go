package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"library/internal/handlers"
	"library/internal/repositories"
	"library/internal/services"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	// Run automatic database migrations
	runMigrations(db)

	// 2. Instantiate data repositories
	bookRepo := repositories.NewBookRepository()
	memberRepo := repositories.NewMemberRepository()
	bookInventoryRepo := repositories.NewBookInventoryRepository()
	libraryRepo := repositories.NewLibraryRepository()
	bookshelfRepo := repositories.NewBookshelfRepository()

	// 3. Inject repositories to bootstrap business workflows
	bookService := services.NewBookService(db, bookRepo, memberRepo, bookInventoryRepo)
	memberService := services.NewMemberService(db, memberRepo, bookRepo)
	bookInventoryService := services.NewBookInventoryService(db, bookInventoryRepo, bookRepo, libraryRepo, bookshelfRepo)
	libraryService := services.NewLibraryService(db, libraryRepo, bookRepo, memberRepo, bookInventoryRepo, bookshelfRepo)
	bookshelfService := services.NewBookshelfService(db, bookshelfRepo, libraryRepo)

	// 4. Bind services into HTTP server payload multiplexer routers
	bookHandler := handlers.NewBookHandler(bookService)
	memberHandler := handlers.NewMemberHandler(memberService)
	bookInventoryHandler := handlers.NewBookInventoryHandler(bookInventoryService, bookshelfService)
	libraryHandler := handlers.NewLibraryHandler(libraryService, bookInventoryService)
	bookshelfHandler := handlers.NewBookshelfHandler(bookshelfService)

	// 5. Register routes on standard ServeMux
	mux := http.NewServeMux()

	// ----------------------------------------------------
	// Libraries
	// ----------------------------------------------------
	mux.HandleFunc("GET /libraries", libraryHandler.ListLibraries)
	mux.HandleFunc("POST /libraries", libraryHandler.RegisterLibrary)
	mux.HandleFunc("GET /libraries/{id}", libraryHandler.GetLibraryByID)
	mux.HandleFunc("PUT /libraries/{id}", libraryHandler.UpdateLibrary)
	mux.HandleFunc("DELETE /libraries/{id}", libraryHandler.DeleteLibrary)

	// ----------------------------------------------------
	// Library Books & Loans
	// ----------------------------------------------------
	mux.HandleFunc("GET /libraries/{id}/books", libraryHandler.GetLibraryBooks)
	mux.HandleFunc("GET /libraries/{id}/loans", libraryHandler.GetLibraryLoans)
	mux.HandleFunc("DELETE /libraries/{id}/books/{book_id}", libraryHandler.DeleteBookFromLibrary)
	mux.HandleFunc("GET /libraries/{id}/books/genres/{genre_id}", libraryHandler.GetLibraryBooksByGenre)
	mux.HandleFunc("POST /libraries/{id}/books/{book_id}/borrow", libraryHandler.BorrowBook)

	// ----------------------------------------------------
	// Bookshelves
	// ----------------------------------------------------
	mux.HandleFunc("GET /libraries/{id}/bookshelves", bookshelfHandler.GetBookshelvesByLibraryID)
	mux.HandleFunc("POST /libraries/{id}/bookshelves", bookshelfHandler.CreateBookshelf)
	mux.HandleFunc("GET /libraries/{id}/bookshelves/{shelf_id}", bookshelfHandler.GetBookshelfByID)
	mux.HandleFunc("DELETE /libraries/{id}/bookshelves/{shelf_id}", bookshelfHandler.DeleteBookshelf)
	mux.HandleFunc("GET /libraries/{id}/bookshelves/{shelf_id}/books", bookshelfHandler.GetBooksByShelfID)

	// ====================================================
	// BOOKS ROUTES
	// ====================================================
	mux.HandleFunc("GET /books", bookHandler.ListBooks)
	mux.HandleFunc("POST /books", bookHandler.CreateBook)
	mux.HandleFunc("GET /books/genres/{id}", bookHandler.GetBooksByGenreID)
	mux.HandleFunc("GET /books/{id}", bookHandler.GetBook)
	mux.HandleFunc("PUT /books/{id}", bookHandler.UpdateBook)
	mux.HandleFunc("DELETE /books/{id}", bookHandler.DeleteBook)

	// ====================================================
	// INVENTORY ROUTES
	// ====================================================
	mux.HandleFunc("GET /inventory", bookInventoryHandler.ListInventory)
	mux.HandleFunc("POST /inventory", bookInventoryHandler.AddInventory)
	mux.HandleFunc("GET /inventory/{libraryId}/{bookId}", bookInventoryHandler.GetAvailableCopies)
	mux.HandleFunc("DELETE /inventory/{libraryId}/{bookId}", bookInventoryHandler.DeleteInventory)

	// ====================================================
	// MEMBERS ROUTES
	// ====================================================
	mux.HandleFunc("GET /members", memberHandler.ListMembers)
	mux.HandleFunc("POST /members", memberHandler.RegisterMember)
	mux.HandleFunc("GET /members/{id}/loans", memberHandler.GetMemberLoans)
	mux.HandleFunc("GET /members/{id}", memberHandler.GetMember)
	mux.HandleFunc("PUT /members/{id}", memberHandler.UpdateMember)
	mux.HandleFunc("DELETE /members/{id}", memberHandler.DeleteMember)

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

func runMigrations(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create migration database driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("Could not initialize migration instance: %v", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Database schema is up to date (no new migrations to run).")
			return
		}
		log.Fatalf("Migration execution failed: %v", err)
	}

	log.Println("Database migrations applied successfully!")
}
