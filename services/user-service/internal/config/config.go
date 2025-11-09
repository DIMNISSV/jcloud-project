// services/user-service/internal/config/config.go
package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string `env:"ENV" env-default:"local"`
	Postgres PostgresConfig
	JWT      JWTConfig
}

type PostgresConfig struct {
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DBName   string `env:"POSTGRES_DB" env-required:"true"`
	Host     string `env:"DB_HOST" env-default:"localhost"`
	Port     string `env:"DB_PORT" env-default:"5432"`
}

type JWTConfig struct {
	Secret string `env:"JWT_SECRET" env-required:"true"`
}

func MustLoad() *Config {
	// Загружаем .env файл только если он существует.
	// В Docker-окружении переменные будут предоставлены через docker-compose.
	if _, err := os.Stat("../../.env"); err == nil {
		if err := godotenv.Load("../../.env"); err != nil {
			log.Fatalf("cannot load .env file: %v", err)
		}
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read config: %v", err)
	}

	// Особая логика для Docker-окружения
	if os.Getenv("DOCKER_ENV") == "true" {
		cfg.Postgres.Host = "db"
	}

	return &cfg
}
