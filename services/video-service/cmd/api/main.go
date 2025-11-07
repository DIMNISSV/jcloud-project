// cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"jcloud-project/video-service/internal/handler"
	"jcloud-project/video-service/internal/repository"
	"jcloud-project/video-service/internal/service"
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
		log.Println("Warning: .env file not found")
	}

	// Database Connection
	dbUser, dbPass, dbName, dbHost := os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"), "localhost"
	if os.Getenv("DOCKER_ENV") == "true" {
		dbHost = "db"
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:5432/%s", dbUser, dbPass, dbHost, dbName)
	dbpool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()
	log.Println("Database connection successful")

	// Dependency Injection
	videoRepo := repository.NewVideoPostgresRepository(dbpool)
	videoService := service.NewVideoService(videoRepo)
	videoHandler := handler.NewVideoHandler(videoService)

	// HTTP Server
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	api := e.Group("/api/v1")

	// JWT Middleware Config
	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return new(service.JwtCustomClaims) },
		SigningKey:    []byte(os.Getenv("JWT_SECRET")),
		ContextKey:    "user",
	}

	// Protected route for video uploads
	api.POST("/videos", videoHandler.UploadVideo, echojwt.WithConfig(jwtConfig))

	e.Logger.Fatal(e.Start(":8081"))
}
