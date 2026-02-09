package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"omnibase/internal/config"
	"omnibase/internal/flow"
	"omnibase/internal/httpapi"
	"omnibase/internal/llm"
	"omnibase/internal/logging"
	"omnibase/internal/mcp"
	"omnibase/internal/qdrant"
)

func main() {
	logger := logging.NewLogger()
	cfg, err := config.Load()
	if err != nil {
		logger.Error("config load failed", "error", err.Error())
		os.Exit(1)
	}

	qdrantClient := qdrant.NewClient(cfg.QdrantURL, cfg.QdrantAPIKey, cfg.QdrantCollection)
	llmClient := llm.NewClient(cfg.LLMBaseURL, cfg.LLMModel)
	toolRegistry := mcp.DefaultTools()
	var sqlExecutor *mcp.SQLExecutor
	if cfg.MySQLDSN != "" {
		sqlExecutor, err = mcp.NewSQLExecutor(cfg.MySQLDriver, cfg.MySQLDSN)
		if err != nil {
			logger.Error("sql executor init failed", "error", err.Error())
			os.Exit(1)
		}
		defer func() {
			if err := sqlExecutor.Close(); err != nil {
				logger.Error("sql executor close failed", "error", err.Error())
			}
		}()
	}
	mcpClient, err := mcp.NewClient(cfg.MCPBaseURL, toolRegistry, sqlExecutor)
	if err != nil {
		logger.Error("mcp client init failed", "error", err.Error())
		os.Exit(1)
	}

	pipeline := flow.Flow{
		Normalizer: flow.RequestNormalizerNode{},
		RAG:        flow.RAGRetrievalNode{LLM: llmClient, Qdrant: qdrantClient, TopK: 5},
		MCP:        flow.MCPToolDispatchNode{Client: mcpClient},
		LLM:        flow.LLMCompletionNode{Client: llmClient, Tools: toolRegistry},
		Formatter:  flow.ResponseFormatterNode{},
	}

	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      httpapi.NewHandler(pipeline, logger),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("omnibase server starting", slog.String("addr", cfg.HTTPAddr))

	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err.Error())
			stop()
		}
	}()

	<-shutdownCtx.Done()
	logger.Info("omnibase server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed", "error", err.Error())
	}
}
