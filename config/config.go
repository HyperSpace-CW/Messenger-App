package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"log"
	"sync"
)

type Config struct {
	Server   ServerConfig
	WSServer WSConfig
	PG       PGConfig
	TokenKey string `env:"TOKEN_KEY,required"`
}

type ServerConfig struct {
	Addr string `env:"SERVER_ADDR,required"`
}

type WSConfig struct {
	Addr string `env:"WS_ADDR,required"`
}

type PGConfig struct {
	Host     string `env:"DB_HOST,required"`
	Port     string `env:"DB_PORT,required"`
	User     string `env:"DB_USER,required"`
	Password string `env:"DB_PASSWORD,required"`
	DBName   string `env:"DB_NAME,required"`
	SSLMode  string `env:"DB_SSLMODE,required"`
}

var (
	config Config
	once   sync.Once
)

func LoadConfig() (*Config, error) {
	once.Do(func() {
		_ = godotenv.Load()
		if err := env.Parse(&config); err != nil {
			log.Fatal(err)
		}
	})
	return &config, nil
}
