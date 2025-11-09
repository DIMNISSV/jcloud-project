// services/billing-service/cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"log"

	"jcloud-project/billing-service/internal/client"
	"jcloud-project/billing-service/internal/config"
	"jcloud-project/billing-service/internal/handler"
	"jcloud-project/billing-service/internal/repository"
	"jcloud-project/billing-service/internal/service"
	commontypes "jcloud-project/libs/go-common/types/jwt"

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
	repo := repository.NewBillingPostgresRepository(dbpool)

	// Create clients
	nextcloudClient := client.NewNextcloudClient(cfg.Nextcloud.ApiURL, cfg.Nextcloud.ApiUser, cfg.Nextcloud.ApiPassword)
	userSvcClient := client.NewUserServiceClient()

	// Inject clients into the service
	billingService := service.NewBillingService(repo, nextcloudClient, userSvcClient)

	handler := handler.NewBillingHandler(billingService)

	//
	// HTTP Server (Echo)
	//
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	//
	// Routes
	//

	// JWT Middleware Config
	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims { return new(commontypes.JwtCustomClaims) },
		SigningKey:    []byte(cfg.JWT.Secret),
		ContextKey:    "user",
	}

	// Public routes
	api := e.Group("/api/v1")
	api.GET("/plans", handler.GetAllPlans)

	// Protected routes
	subscriptionsAPI := api.Group("/subscriptions")
	subscriptionsAPI.Use(echojwt.WithConfig(jwtConfig))
	subscriptionsAPI.GET("/me", handler.GetUserSubscription)
	subscriptionsAPI.POST("", handler.ChangeSubscription)

	// Internal routes
	internalAPI := e.Group("/internal/v1")
	internalAPI.GET("/permissions/:userId", handler.GetUserPermissions)
	internalAPI.POST("/subscriptions", handler.CreateSubscription)

	// Start server
	log.Println("Starting billing-service on :8082")
	e.Logger.Fatal(e.Start(":8082"))
}
