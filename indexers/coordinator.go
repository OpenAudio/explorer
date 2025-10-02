package indexers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/OpenAudio/go-openaudio/pkg/sdk"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Coordinator manages and runs all indexers
type Coordinator struct {
	pool           *pgxpool.Pool
	sdk            *sdk.OpenAudioSDK
	logger         *zap.Logger
	indexers       []Indexer
	startingBlock  int64
	checkReadiness bool
	statusMu       sync.RWMutex
}

// NewCoordinator creates a new indexer coordinator
func NewCoordinator(
	pool *pgxpool.Pool,
	sdk *sdk.OpenAudioSDK,
	logger *zap.Logger,
	startingBlock int64,
	checkReadiness bool,
) *Coordinator {
	return &Coordinator{
		pool:           pool,
		sdk:            sdk,
		logger:         logger,
		startingBlock:  startingBlock,
		checkReadiness: checkReadiness,
		indexers:       make([]Indexer, 0),
	}
}

// Initialize sets up all indexers
func (c *Coordinator) Initialize() {
	// Create blocks indexer (RPC -> Postgres)
	blocksIndexer := NewBlocksIndexer(c.pool, c.sdk, c.logger, c.startingBlock, c.checkReadiness)
	c.indexers = append(c.indexers, blocksIndexer)

	// Create downstream indexers (Postgres -> Postgres)
	// These read from the blocks/transactions tables
	playsIndexer := NewPlaysIndexer(c.pool, c.logger)
	c.indexers = append(c.indexers, playsIndexer)

	validatorsIndexer := NewValidatorsIndexer(c.pool, c.logger)
	c.indexers = append(c.indexers, validatorsIndexer)

	// TODO: Add more indexers as they are implemented
	// manageEntityIndexer := NewManageEntityIndexer(c.pool, c.logger)
	// c.indexers = append(c.indexers, manageEntityIndexer)

	// slaIndexer := NewSLARollupIndexer(c.pool, c.logger)
	// c.indexers = append(c.indexers, slaIndexer)

	// storageProofsIndexer := NewStorageProofsIndexer(c.pool, c.logger)
	// c.indexers = append(c.indexers, storageProofsIndexer)

	c.logger.Info("Initialized indexers", zap.Int("count", len(c.indexers)))
}

// Start runs all indexers concurrently in a wait group
func (c *Coordinator) Start(ctx context.Context) error {
	if len(c.indexers) == 0 {
		c.Initialize()
	}

	c.logger.Info("Starting indexer coordinator",
		zap.Int("indexerCount", len(c.indexers)),
		zap.Int64("startingBlock", c.startingBlock))

	// Use errgroup to manage all indexer goroutines
	wg, ctx := errgroup.WithContext(ctx)

	// Start status reporter
	wg.Go(func() error {
		return c.statusReporter(ctx)
	})

	// Start each indexer in its own goroutine
	for _, indexer := range c.indexers {
		idx := indexer // Capture for closure

		wg.Go(func() error {
			c.logger.Info("Starting indexer",
				zap.String("name", idx.Name()))

			// Add retry logic with exponential backoff
			retryCount := 0
			maxRetries := 5

			for {
				err := idx.Start(ctx)
				if err == nil {
					return nil
				}

				// Check if context is cancelled
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				if retryCount >= maxRetries {
					return fmt.Errorf("indexer %s failed after %d retries: %w",
						idx.Name(), maxRetries, err)
				}

				retryCount++
				backoff := time.Duration(retryCount*retryCount) * time.Second
				c.logger.Error("Indexer failed, retrying",
					zap.String("name", idx.Name()),
					zap.Error(err),
					zap.Int("retry", retryCount),
					zap.Duration("backoff", backoff))

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
					// Continue to retry
				}
			}
		})
	}

	// Wait for all indexers
	if err := wg.Wait(); err != nil {
		c.logger.Error("Coordinator error", zap.Error(err))
		return err
	}

	return nil
}

// statusReporter periodically logs the status of all indexers
func (c *Coordinator) statusReporter(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			status, err := c.GetStatus()
			if err != nil {
				c.logger.Warn("Failed to get status", zap.Error(err))
				continue
			}

			for name, state := range status {
				c.logger.Info("Indexer status",
					zap.String("indexer", name),
					zap.Int64("lastBlock", state.LastIndexedBlock),
					zap.Int64("targetBlock", state.TargetBlock),
					zap.Int64("remaining", state.BlocksRemaining),
					zap.Float64("progress", state.PercentageComplete),
					zap.String("status", state.Status))
			}
		}
	}
}

// GetStatus returns the current status of all indexers
func (c *Coordinator) GetStatus() (map[string]*IndexerState, error) {
	status := make(map[string]*IndexerState)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errs []error

	for _, indexer := range c.indexers {
		wg.Add(1)
		go func(idx Indexer) {
			defer wg.Done()
			state, err := idx.GetState()
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", idx.Name(), err))
				mu.Unlock()
				return
			}
			mu.Lock()
			status[idx.Name()] = state
			mu.Unlock()
		}(indexer)
	}

	wg.Wait()

	if len(errs) > 0 {
		// Return partial status even if some indexers failed
		c.logger.Warn("Some indexers failed to report status",
			zap.Int("errors", len(errs)))
	}

	return status, nil
}

// GetIndexer returns a specific indexer by name
func (c *Coordinator) GetIndexer(name string) Indexer {
	for _, idx := range c.indexers {
		if idx.Name() == name {
			return idx
		}
	}
	return nil
}

// Stop gracefully stops all indexers
func (c *Coordinator) Stop() {
	c.logger.Info("Stopping indexer coordinator")
	// Context cancellation will stop all indexers
}