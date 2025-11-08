// cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"jcloud-project/billing-service/internal/handler"
	"jcloud-project/billing-service/internal/repository"
	"jcloud-project/billing-service/internal/service"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Info: .env file not found, relying on environment variables.")
	}

	// Configuration
	dbUser, dbPass, dbName, dbHost := os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"), "localhost"
	if os.Getenv("DOCKER_ENV") == "true" {
		dbHost = "db"
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	// Database Connection
	connStr := fmt.Sprintf("postgres://%s:%s@%s:5432/%s", dbUser, dbPass, dbHost, dbName)
	dbpool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()
	log.Println("Database connection successful")

	// Dependency Injection
	repo := repository.NewBillingPostgresRepository(dbpool)
	billingService := service.NewBillingService(repo)
	handler := handler.NewBillingHandler(billingService)

	// HTTP Server
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	//
	// Routes
	//

	// JWT Middleware Config
	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return new(service.JwtCustomClaims) },
		SigningKey:    []byte(jwtSecret),
		ContextKey:    "user",
	}

	// Public routes
	api := e.Group("/api/v1")
	api.GET("/plans", handler.GetAllPlans)

	// Protected routes
	subscriptionsAPI := api.Group("/subscriptions")
	subscriptionsAPI.Use(echojwt.WithConfig(jwtConfig))
	subscriptionsAPI.GET("/me", handler.GetUserSubscription)

	// Internal routes
	internalAPI := e.Group("/internal/v1")
	internalAPI.GET("/permissions/:userId", handler.GetUserPermissions)
	internalAPI.POST("/subscriptions", handler.CreateSubscription)

	// Start server
	log.Println("Starting billing-service on :8082")
	e.Logger.Fatal(e.Start(":8082"))
}
