// services/user-service/cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"log"

	commontypes "jcloud-project/libs/go-common/types/jwt"
	"jcloud-project/user-service/internal/config"
	"jcloud-project/user-service/internal/handler"
	"jcloud-project/user-service/internal/repository"
	"jcloud-project/user-service/internal/service"

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
	userRepo := repository.NewUserPostgresRepository(dbpool)
	// Передаем JWT-секрет в сервис, где он действительно нужен
	userService := service.NewUserService(userRepo, cfg.JWT.Secret)
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
	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return new(commontypes.JwtCustomClaims) },
		SigningKey:    []byte(cfg.JWT.Secret),
		ContextKey:    "user",
	}

	// Admin routes - requires both JWT auth and Admin role
	adminAPI := api.Group("/admin")
	adminAPI.Use(echojwt.WithConfig(jwtConfig))
	adminAPI.Use(handler.AdminMiddleware) // Наша кастомная middleware для проверки роли ADMIN

	adminAPI.GET("/users", userHandler.GetAllUsers)
	adminAPI.PATCH("/users/:userId", userHandler.PatchUser)

	// Internal routes for service-to-service communication
	internalAPI := e.Group("/internal/v1")
	internalAPI.GET("/users/:userId", userHandler.GetInternalUserDetails)

	// Start server
	log.Println("Starting user-service on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
