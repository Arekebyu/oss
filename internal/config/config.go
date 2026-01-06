package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	ElasticsearchURL string
	MLServiceAddr    string
	Port             string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://admin:secretpassword@localhost:5432/search_engine"),
		ElasticsearchURL: getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		MLServiceAddr:    getEnv("ML_SERVICE_ADDR", "localhost:50051"),
		Port:             getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
