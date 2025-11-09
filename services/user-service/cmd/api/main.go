// services/user-service/cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"jcloud-project/user-service/internal/handler"
	"jcloud-project/user-service/internal/repository"
	"jcloud-project/user-service/internal/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load .env file for local development.
	// In Docker, environment variables are provided by docker-compose.
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Info: .env file not found, relying on environment variables.")
	}

	//
	// Configuration is read from environment variables
	//
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	jwtSecret := os.Getenv("JWT_SECRET")

	dbHost := "localhost" // Default for local run
	if os.Getenv("DOCKER_ENV") == "true" {
		dbHost = "db" // Docker internal network hostname
	}

	//
	// Database Connection
	//
	connStr := fmt.Sprintf("postgres://%s:%s@%s:5432/%s", dbUser, dbPass, dbHost, dbName)
	dbpool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()
	log.Println("Database connection successful")

	//
	// Dependency Injection
	//
	userRepo := repository.NewUserPostgresRepository(dbpool)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	//
	// HTTP Server (Echo)
	//
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	//
	// Routes
	//
	api := e.Group("/api/v1")

	// Public routes
	api.POST("/users/register", userHandler.Register)
	api.POST("/users/login", userHandler.Login)

	// Protected routes
	protected := api.Group("/users")
	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return new(service.JwtCustomClaims) },
		SigningKey:    []byte(jwtSecret),
		ContextKey:    "user",
	}
	protected.Use(echojwt.WithConfig(jwtConfig))

	protected.GET("/me", userHandler.Profile)

	// Internal routes for service-to-service communication
	internalAPI := e.Group("/internal/v1")
	internalAPI.GET("/users/:userId", userHandler.GetInternalUserDetails)

	// Admin routes - requires both JWT auth and Admin role
	adminAPI := api.Group("/admin")
	adminAPI.Use(echojwt.WithConfig(jwtConfig)) // 1. Must be a valid user
	adminAPI.Use(handler.AdminMiddleware)       // 2. Must be an admin

	adminAPI.GET("/users", userHandler.GetAllUsers)
	adminAPI.PUT("/users/:userId", userHandler.UpdateUser)

	// Start server
	log.Println("Starting user-service on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
