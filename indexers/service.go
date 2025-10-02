package indexers

import (
	"context"
	"fmt"
	"sync"

	sdk "github.com/OpenAudio/go-openaudio/pkg/sdk"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Service coordinates all indexers
type Service struct {
	pool           *pgxpool.Pool
	sdk            *sdk.OpenAudioSDK
	logger         *zap.Logger
	indexers       []Indexer
	startingBlock  int64
	checkReadiness bool
}

// NewService creates a new indexer service
func NewService(
	pool *pgxpool.Pool,
	sdk *sdk.OpenAudioSDK,
	logger *zap.Logger,
	startingBlock int64,
	checkReadiness bool,
) *Service {
	return &Service{
		pool:           pool,
		sdk:            sdk,
		logger:         logger,
		startingBlock:  startingBlock,
		checkReadiness: checkReadiness,
		indexers:       make([]Indexer, 0),
	}
}

// Initialize sets up all indexers
func (s *Service) Initialize() {
	// Create blocks indexer (RPC -> Postgres)
	blocksIndexer := NewBlocksIndexer(s.pool, s.sdk, s.logger, s.startingBlock, s.checkReadiness)
	s.indexers = append(s.indexers, blocksIndexer)

	// Create downstream indexers (Postgres -> Postgres)
	playsIndexer := NewPlaysIndexer(s.pool, s.logger)
	s.indexers = append(s.indexers, playsIndexer)

	validatorsIndexer := NewValidatorsIndexer(s.pool, s.logger)
	s.indexers = append(s.indexers, validatorsIndexer)

	// TODO: Add more indexers as needed
	// manageEntityIndexer := NewManageEntityIndexer(s.pool, s.logger)
	// s.indexers = append(s.indexers, manageEntityIndexer)

	// slaIndexer := NewSLAIndexer(s.pool, s.logger)
	// s.indexers = append(s.indexers, slaIndexer)

	// storageProofsIndexer := NewStorageProofsIndexer(s.pool, s.logger)
	// s.indexers = append(s.indexers, storageProofsIndexer)

	s.logger.Info("Initialized indexers", zap.Int("count", len(s.indexers)))
}

// Start runs all indexers concurrently
func (s *Service) Start(ctx context.Context) error {
	if len(s.indexers) == 0 {
		s.Initialize()
	}

	s.logger.Info("Starting indexer service", zap.Int("indexerCount", len(s.indexers)))

	// Use errgroup to manage all indexer goroutines
	g, ctx := errgroup.WithContext(ctx)

	// Start each indexer
	for _, indexer := range s.indexers {
		idx := indexer // Capture for closure
		g.Go(func() error {
			s.logger.Info("Starting indexer", zap.String("name", idx.Name()))
			if err := idx.Start(ctx); err != nil {
				return fmt.Errorf("indexer %s failed: %w", idx.Name(), err)
			}
			return nil
		})
	}

	// Wait for all indexers
	if err := g.Wait(); err != nil {
		s.logger.Error("Indexer service error", zap.Error(err))
		return err
	}

	return nil
}

// GetStatus returns the status of all indexers
func (s *Service) GetStatus() (map[string]*IndexerState, error) {
	status := make(map[string]*IndexerState)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, indexer := range s.indexers {
		wg.Add(1)
		go func(idx Indexer) {
			defer wg.Done()
			state, err := idx.GetState()
			if err != nil {
				s.logger.Warn("Failed to get indexer state",
					zap.String("name", idx.Name()),
					zap.Error(err))
				return
			}
			mu.Lock()
			status[idx.Name()] = state
			mu.Unlock()
		}(indexer)
	}

	wg.Wait()
	return status, nil
}

// Stop gracefully stops all indexers
func (s *Service) Stop() {
	s.logger.Info("Stopping indexer service")
	// Context cancellation will stop all indexers
}
