// services/video-service/cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"log"

	"jcloud-project/video-service/internal/config"
	"jcloud-project/video-service/internal/handler"
	"jcloud-project/video-service/internal/repository"
	"jcloud-project/video-service/internal/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	//
	// Configuration
	//
	cfg := config.MustLoad()

	//
	// Database Connection
	//
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName)
	dbpool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()
	log.Println("Database connection successful")

	//
	// Dependency Injection
	//
	videoRepo := repository.NewVideoPostgresRepository(dbpool)
	videoService := service.NewVideoService(videoRepo)
	videoHandler := handler.NewVideoHandler(videoService)

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

	// JWT Middleware Config
	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return new(service.JwtCustomClaims) },
		SigningKey:    []byte(cfg.JWT.Secret),
		ContextKey:    "user",
	}

	// Protected route for video uploads
	api.POST("/videos", videoHandler.UploadVideo, echojwt.WithConfig(jwtConfig))

	// Start server
	log.Println("Starting video-service on :8081")
	e.Logger.Fatal(e.Start(":8081"))
}
