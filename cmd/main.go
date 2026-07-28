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
	"library/internal/middleware"
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
	userRepo := repositories.NewUserRepository()
	empRepo := repositories.NewEmployeeRepository()

	// 3. Inject repositories to bootstrap business workflows
	bookService := services.NewBookService(db, bookRepo, memberRepo, bookInventoryRepo)
	memberService := services.NewMemberService(db, memberRepo, bookRepo)
	bookInventoryService := services.NewBookInventoryService(db, bookInventoryRepo, bookRepo, libraryRepo, bookshelfRepo)
	libraryService := services.NewLibraryService(db, libraryRepo, bookRepo, memberRepo, bookInventoryRepo, bookshelfRepo)
	bookshelfService := services.NewBookshelfService(db, bookshelfRepo, libraryRepo)
	userService := services.NewUserService(db, userRepo)
	empService := services.NewEmployeeService(db, empRepo, memberRepo, userRepo, libraryRepo)

	// 4. Bind services into HTTP server payload multiplexer routers
	bookHandler := handlers.NewBookHandler(bookService)
	memberHandler := handlers.NewMemberHandler(memberService)
	bookInventoryHandler := handlers.NewBookInventoryHandler(bookInventoryService, bookshelfService)
	libraryHandler := handlers.NewLibraryHandler(libraryService, bookInventoryService)
	bookshelfHandler := handlers.NewBookshelfHandler(bookshelfService)
	userHandler := handlers.NewUserHandler(userService)
	empHandler := handlers.NewEmployeeHandler(empService)

	//middleware
	authMiddleware := middleware.Authenticate(userService)

	// 5. Register routes on standard ServeMux
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", userHandler.Register)
	mux.HandleFunc("POST /login", userHandler.Login)
	mux.Handle("POST /register_client", protected(authMiddleware, userHandler.Register, "admin", "employee"))

	// ----------------------------------------------------
	// Libraries
	// ----------------------------------------------------
	mux.Handle("GET /libraries", protected(authMiddleware, libraryHandler.ListLibraries))
	mux.Handle("POST /libraries", protected(authMiddleware, libraryHandler.RegisterLibrary, "admin"))
	mux.Handle("GET /libraries/{id}", protected(authMiddleware, libraryHandler.GetLibraryByID))
	mux.Handle("PUT /libraries/{id}", protected(authMiddleware, libraryHandler.UpdateLibrary, "admin"))
	mux.Handle("DELETE /libraries/{id}", protected(authMiddleware, libraryHandler.DeleteLibrary, "admin"))

	// ----------------------------------------------------
	// Library Books & Loans
	// ----------------------------------------------------
	mux.Handle("GET /libraries/{id}/books", protected(authMiddleware, libraryHandler.GetLibraryBooks))
	mux.Handle("GET /libraries/{id}/loans", protected(authMiddleware, libraryHandler.GetLibraryLoans, "admin", "employee"))
	mux.Handle("DELETE /libraries/{id}/books/{book_id}", protected(authMiddleware, libraryHandler.DeleteBookFromLibrary, "admin"))
	mux.Handle("GET /libraries/{id}/books/genres/{genre_id}", protected(authMiddleware, libraryHandler.GetLibraryBooksByGenre))
	mux.Handle("POST /libraries/{id}/books/{book_id}/borrow", protected(authMiddleware, libraryHandler.BorrowBook, "client"))
	mux.Handle("POST /libraries/{id}/return", protected(authMiddleware, libraryHandler.ReturnBook, "client"))
	mux.Handle("GET /libraries/{id}/returned_books", protected(authMiddleware, libraryHandler.ListReturnedBooks, "employee"))
	mux.Handle("POST /libraries/{id}/returned_books/{book_id}/assign_shelf", protected(authMiddleware, libraryHandler.AssignShelf, "employee"))

	// ----------------------------------------------------
	// Bookshelves
	// ----------------------------------------------------
	mux.Handle("GET /libraries/{id}/bookshelves", protected(authMiddleware, bookshelfHandler.GetBookshelvesByLibraryID))
	mux.Handle("POST /libraries/{id}/bookshelves", protected(authMiddleware, bookshelfHandler.CreateBookshelf, "admin"))
	mux.Handle("GET /libraries/{id}/bookshelves/{shelf_id}", protected(authMiddleware, bookshelfHandler.GetBookshelfByID))
	mux.Handle("DELETE /libraries/{id}/bookshelves/{shelf_id}", protected(authMiddleware, bookshelfHandler.DeleteBookshelf, "admin"))
	mux.Handle("GET /libraries/{id}/bookshelves/{shelf_id}/books", protected(authMiddleware, bookshelfHandler.GetBooksByShelfID))

	// ====================================================
	// BOOKS ROUTES (All require Auth, Admin restricted where needed)
	// ====================================================

	// 1. Authenticated User Routes (Any logged-in user can view/list)
	mux.Handle("GET /books", protected(authMiddleware, bookHandler.ListBooks))
	mux.Handle("GET /books/genres/{id}", protected(authMiddleware, bookHandler.GetBooksByGenreID))
	mux.Handle("GET /books/{id}", protected(authMiddleware, bookHandler.GetBook))

	//admin
	mux.Handle("POST /books", protected(authMiddleware, bookHandler.CreateBook, "admin"))
	mux.Handle("PUT /books/{id}", protected(authMiddleware, bookHandler.UpdateBook, "admin"))
	mux.Handle("DELETE /books/{id}", protected(authMiddleware, bookHandler.DeleteBook, "admin"))

	// ====================================================
	// INVENTORY ROUTES (Admin-Only)
	// ====================================================

	mux.Handle("GET /inventory", protected(authMiddleware, bookInventoryHandler.ListInventory, "admin"))
	mux.Handle("POST /inventory", protected(authMiddleware, bookInventoryHandler.AddInventory, "admin"))
	mux.Handle("GET /inventory/{libraryId}/{bookId}", protected(authMiddleware, bookInventoryHandler.GetAvailableCopies, "admin"))
	mux.Handle("DELETE /inventory/{libraryId}/{bookId}", protected(authMiddleware, bookInventoryHandler.DeleteInventory, "admin"))

	// ====================================================
	// MEMBER ROUTES (Admin-Only)
	// ====================================================

	mux.Handle("GET /clients", protected(authMiddleware, memberHandler.ListMembers, "admin"))
	mux.Handle("POST /clients", protected(authMiddleware, userHandler.RegisterClient, "admin"))
	mux.Handle("GET /clients/{id}", protected(authMiddleware, memberHandler.GetMember, "admin"))
	mux.Handle("GET /clients/{id}/loans", protected(authMiddleware, memberHandler.GetMemberLoans, "admin"))
	mux.Handle("PUT /clients/{id}", protected(authMiddleware, memberHandler.UpdateMember, "admin"))
	mux.Handle("DELETE /clients/{id}", protected(authMiddleware, memberHandler.DeleteMember, "admin"))

	// Protected routes (Admin only)
	mux.Handle("GET /employees", protected(authMiddleware, empHandler.ListEmployees, "admin"))
	mux.Handle("GET /employees/{member_id}", protected(authMiddleware, empHandler.GetEmployee, "admin"))
	mux.Handle("POST /employees", protected(authMiddleware, empHandler.RegisterEmployee, "admin"))
	mux.Handle("PUT /employees/{member_id}", protected(authMiddleware, empHandler.UpdateEmployee, "admin"))
	mux.Handle("DELETE /employees/{member_id}", protected(authMiddleware, empHandler.DeleteEmployee, "admin"))

	mux.Handle("/loans", protected(authMiddleware, bookHandler.HandleListLoans, "admin"))

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

// Helper to apply middleware to multiple routes cleanly
func protected(authMiddleware func(http.Handler) http.Handler, handler http.HandlerFunc, roles ...string) http.Handler {
	if len(roles) > 0 {
		return middleware.Chain(handler, authMiddleware, middleware.RequireRoles(roles...))
	}
	return middleware.Chain(handler, authMiddleware)
}
