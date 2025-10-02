package indexers

import (
	"context"
	"fmt"
	"time"

	"github.com/OpenAudio/explorer/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Indexer is the interface that all indexers must implement
type Indexer interface {
	// Name returns the unique name of this indexer
	Name() string

	// Start begins the indexing process
	Start(ctx context.Context) error

	// GetState returns the current indexer state
	GetState() (*IndexerState, error)
}

// IndexerState represents the state of an indexer
type IndexerState struct {
	Name               string
	LastIndexedBlock   int64
	TargetBlock        int64
	Status             string
	ErrorMessage       *string
	BatchSize          int32
	LastRunAt          *time.Time
	BlocksRemaining    int64
	PercentageComplete float64
}

// BaseIndexer provides common functionality for all indexers
type BaseIndexer struct {
	name       string
	pool       *pgxpool.Pool
	queries    *db.Queries
	logger     *zap.Logger
	batchSize  int32
	maxRetries int
}

// NewBaseIndexer creates a new base indexer
func NewBaseIndexer(name string, pool *pgxpool.Pool, logger *zap.Logger, batchSize int32) *BaseIndexer {
	return &BaseIndexer{
		name:       name,
		pool:       pool,
		queries:    db.New(pool),
		logger:     logger.With(zap.String("indexer", name)),
		batchSize:  batchSize,
		maxRetries: 3,
	}
}

// Name returns the name of the indexer
func (b *BaseIndexer) Name() string {
	return b.name
}

// GetState retrieves the current state from the database
func (b *BaseIndexer) GetState() (*IndexerState, error) {
	ctx := context.Background()
	dbState, err := b.queries.GetIndexerState(ctx, b.name)
	if err != nil {
		return nil, fmt.Errorf("failed to get indexer state: %w", err)
	}

	blocksRemaining := dbState.TargetBlock - dbState.LastIndexedBlock
	if blocksRemaining < 0 {
		blocksRemaining = 0
	}

	percentageComplete := float64(0)
	if dbState.TargetBlock > 0 {
		percentageComplete = float64(dbState.LastIndexedBlock) / float64(dbState.TargetBlock) * 100
		if percentageComplete > 100 {
			percentageComplete = 100
		}
	}

	// Convert pgtype fields to Go types
	var errorMessage *string
	if dbState.ErrorMessage.Valid {
		errorMessage = &dbState.ErrorMessage.String
	}

	var lastRunAt *time.Time
	if dbState.LastRunAt.Valid {
		lastRunAt = &dbState.LastRunAt.Time
	}

	return &IndexerState{
		Name:               dbState.IndexerName,
		LastIndexedBlock:   dbState.LastIndexedBlock,
		TargetBlock:        dbState.TargetBlock,
		Status:             dbState.Status,
		ErrorMessage:       errorMessage,
		BatchSize:          dbState.BatchSize,
		LastRunAt:          lastRunAt,
		BlocksRemaining:    blocksRemaining,
		PercentageComplete: percentageComplete,
	}, nil
}

// InitializeState ensures the indexer state exists in the database
func (b *BaseIndexer) InitializeState(ctx context.Context, startBlock int64) error {
	now := time.Now()
	return b.queries.UpsertIndexerState(ctx, db.UpsertIndexerStateParams{
		IndexerName:      b.name,
		LastIndexedBlock: startBlock,
		TargetBlock:      startBlock,
		Status:           "idle",
		ErrorMessage:     pgtype.Text{Valid: false},
		BatchSize:        b.batchSize,
		LastRunAt:        pgtype.Timestamp{Time: now, Valid: true},
	})
}

// UpdateProgress updates the indexer's progress
func (b *BaseIndexer) UpdateProgress(ctx context.Context, lastBlock int64, status string) error {
	return b.queries.UpdateIndexerProgress(ctx, db.UpdateIndexerProgressParams{
		IndexerName:      b.name,
		LastIndexedBlock: lastBlock,
		Status:           status,
	})
}

// UpdateTarget updates the target block for this indexer
func (b *BaseIndexer) UpdateTarget(ctx context.Context, targetBlock int64) error {
	state, err := b.queries.GetIndexerState(ctx, b.name)
	if err != nil {
		return err
	}

	now := time.Now()
	return b.queries.UpsertIndexerState(ctx, db.UpsertIndexerStateParams{
		IndexerName:      b.name,
		LastIndexedBlock: state.LastIndexedBlock,
		TargetBlock:      targetBlock,
		Status:           state.Status,
		ErrorMessage:     state.ErrorMessage,
		BatchSize:        state.BatchSize,
		LastRunAt:        pgtype.Timestamp{Time: now, Valid: true},
	})
}

// SetError sets an error state for the indexer
func (b *BaseIndexer) SetError(ctx context.Context, err error) error {
	errMsg := err.Error()
	return b.queries.UpdateIndexerError(ctx, db.UpdateIndexerErrorParams{
		IndexerName:  b.name,
		ErrorMessage: pgtype.Text{String: errMsg, Valid: true},
	})
}

// GetLatestBlockHeight gets the latest block height from the blocks table
func (b *BaseIndexer) GetLatestBlockHeight(ctx context.Context) (int64, error) {
	var height int64
	err := b.pool.QueryRow(ctx, "SELECT COALESCE(MAX(height), 0) FROM blocks").Scan(&height)
	if err != nil {
		return 0, err
	}
	return height, nil
}

// ShouldUseBatchMode determines if the indexer should batch process
func (b *BaseIndexer) ShouldUseBatchMode(blocksRemaining int64) bool {
	// Use batch mode if we're more than 10 batches behind
	return blocksRemaining > int64(b.batchSize)*10
}

// CalculateBatchSize dynamically adjusts batch size based on how far behind we are
func (b *BaseIndexer) CalculateBatchSize(blocksRemaining int64) int32 {
	if blocksRemaining > 10000 {
		return b.batchSize * 10 // 10x batch size when very far behind
	} else if blocksRemaining > 1000 {
		return b.batchSize * 5 // 5x batch size when moderately behind
	} else if blocksRemaining > 100 {
		return b.batchSize * 2 // 2x batch size when slightly behind
	}
	return b.batchSize // Normal batch size
}

// Sleep with context cancellation support
func (b *BaseIndexer) Sleep(ctx context.Context, duration time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(duration):
		return nil
	}
}
