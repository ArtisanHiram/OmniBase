package config

import (
	"errors"
	"os"
)

type Config struct {
	HTTPAddr         string
	QdrantURL        string
	QdrantAPIKey     string
	QdrantCollection string
	LLMBaseURL       string
	LLMModel         string
	MCPBaseURL       string
	MySQLDriver      string
	MySQLDSN         string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:         getenvDefault("OMNIBASE_HTTP_ADDR", ":8080"),
		QdrantURL:        getenvDefault("OMNIBASE_QDRANT_URL", "http://localhost:6333"),
		QdrantAPIKey:     os.Getenv("OMNIBASE_QDRANT_API_KEY"),
		QdrantCollection: getenvDefault("OMNIBASE_QDRANT_COLLECTION", "omnibase_docs"),
		LLMBaseURL:       getenvDefault("OMNIBASE_LLM_BASE_URL", "http://localhost:8000"),
		LLMModel:         getenvDefault("OMNIBASE_LLM_MODEL", "qwen2.5-coder-14b"),
		MCPBaseURL:       getenvDefault("OMNIBASE_MCP_BASE_URL", "http://localhost:7000"),
		MySQLDriver:      getenvDefault("OMNIBASE_MYSQL_DRIVER", "mysql"),
		MySQLDSN:         os.Getenv("OMNIBASE_MYSQL_DSN"),
	}

	if cfg.HTTPAddr == "" {
		return Config{}, errors.New("OMNIBASE_HTTP_ADDR is required")
	}

	return cfg, nil
}

func getenvDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
