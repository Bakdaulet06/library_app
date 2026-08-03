package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"library/internal/geocoder"
	"library/internal/handlers"
	"library/internal/middleware"
	"library/internal/repositories"
	"library/internal/routes"
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

	geocoder := geocoder.NewGeocoder()

	// 2. Instantiate data repositories
	bookRepo := repositories.NewBookRepository()
	memberRepo := repositories.NewMemberRepository()
	bookInventoryRepo := repositories.NewBookInventoryRepository()
	libraryRepo := repositories.NewLibraryRepository()
	bookshelfRepo := repositories.NewBookshelfRepository()
	userRepo := repositories.NewUserRepository()
	empRepo := repositories.NewEmployeeRepository()
	orderRepo := repositories.NewOrderRepository()
	profileRepo := repositories.NewProfileRepository()
	genreRepo := repositories.NewGenreRepository()

	// 3. Inject repositories to bootstrap business workflows
	bookService := services.NewBookService(db, bookRepo, memberRepo, bookInventoryRepo)
	memberService := services.NewMemberService(db, memberRepo, bookRepo)
	bookInventoryService := services.NewBookInventoryService(db, bookInventoryRepo, bookRepo, libraryRepo, bookshelfRepo)
	libraryService := services.NewLibraryService(db, libraryRepo, bookRepo, memberRepo, bookInventoryRepo, bookshelfRepo, *geocoder, profileRepo)
	bookshelfService := services.NewBookshelfService(db, bookshelfRepo, libraryRepo)
	userService := services.NewUserService(db, userRepo)
	empService := services.NewEmployeeService(db, empRepo, memberRepo, userRepo, libraryRepo)
	orderService := services.NewOrderService(db, orderRepo, bookInventoryRepo, bookshelfRepo)
	profileService := services.NewProfileService(db, profileRepo)
	genreService := services.NewGenreService(db, genreRepo)

	// 4. Bind services into HTTP server payload multiplexer routers
	bookHandler := handlers.NewBookHandler(bookService)
	memberHandler := handlers.NewMemberHandler(memberService)
	bookInventoryHandler := handlers.NewBookInventoryHandler(bookInventoryService, bookshelfService)
	libraryHandler := handlers.NewLibraryHandler(libraryService, bookInventoryService)
	bookshelfHandler := handlers.NewBookshelfHandler(bookshelfService)
	userHandler := handlers.NewUserHandler(userService)
	empHandler := handlers.NewEmployeeHandler(empService)
	orderHandler := handlers.NewOrderHandler(orderService, profileService)
	profileHandler := handlers.NewProfileHandler(profileService)
	genreHandler := handlers.NewGenreHandler(genreService)

	// middleware
	authMiddleware := middleware.Authenticate(userService)

	// 5. Register routes on standard ServeMux, grouped in internal/routes
	mux := http.NewServeMux()

	routes.RegisterAll(mux, routes.Dependencies{
		BookHandler:          bookHandler,
		MemberHandler:        memberHandler,
		BookInventoryHandler: bookInventoryHandler,
		LibraryHandler:       libraryHandler,
		BookshelfHandler:     bookshelfHandler,
		UserHandler:          userHandler,
		EmployeeHandler:      empHandler,
		AuthMiddleware:       authMiddleware,
		OrderHandler:         orderHandler,
		ProfileHandler:       profileHandler,
		GenreHandler:         genreHandler,
	})

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
