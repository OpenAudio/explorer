package main

import (
	"context"
	"fmt"
	"log"

	"github.com/OpenAudio/explorer/config"
	"github.com/OpenAudio/explorer/indexers"
	"github.com/OpenAudio/explorer/server"
	sdk "github.com/OpenAudio/go-openaudio/pkg/sdk"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	cfg := config.NewConfig()

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	defer logger.Sync()

	// Create database connection pool
	pool, err := pgxpool.New(context.Background(), cfg.PgURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	// Create SDK client
	oap := sdk.NewOpenAudioSDK(cfg.NodeURL)

	// Create server
	s := server.New(cfg)

	// Create indexer coordinator
	coordinator := indexers.NewCoordinator(
		pool,
		oap,
		logger,
		1,     // starting block
		false, // check readiness
	)

	// Start everything in an errgroup
	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)

	// Start server
	g.Go(func() error {
		if err := s.Start(); err != nil {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	})

	// Start indexers
	g.Go(func() error {
		if err := coordinator.Start(ctx); err != nil {
			return fmt.Errorf("indexer error: %w", err)
		}
		return nil
	})

	// Wait for all components
	if err := g.Wait(); err != nil {
		logger.Fatal("Application error", zap.Error(err))
	}
}
