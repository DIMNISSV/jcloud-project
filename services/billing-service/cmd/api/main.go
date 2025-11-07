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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
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
	service := service.NewBillingService(repo)
	handler := handler.NewBillingHandler(service)

	// HTTP Server
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Internal routes - not exposed to the public internet via API Gateway
	internalAPI := e.Group("/internal/v1")
	internalAPI.GET("/permissions/:userId", handler.GetUserPermissions)
	internalAPI.POST("/subscriptions", handler.CreateSubscription)

	// Start server
	log.Println("Starting billing-service on :8082")
	e.Logger.Fatal(e.Start(":8082"))
}
