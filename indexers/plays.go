package indexers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/OpenAudio/explorer/db"
	corev1 "github.com/OpenAudio/go-openaudio/pkg/api/core/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

// PlaysIndexer indexes play transactions from the transactions table
type PlaysIndexer struct {
	*BaseIndexer
	pollInterval      time.Duration
	batchPollInterval time.Duration
}

// NewPlaysIndexer creates a new plays indexer
func NewPlaysIndexer(pool *pgxpool.Pool, logger *zap.Logger) *PlaysIndexer {
	return &PlaysIndexer{
		BaseIndexer:       NewBaseIndexer("plays", pool, logger, 100),
		pollInterval:      2 * time.Second,
		batchPollInterval: 200 * time.Millisecond,
	}
}

// Start begins the plays indexing process
func (pi *PlaysIndexer) Start(ctx context.Context) error {
	// Initialize state if needed
	state, err := pi.queries.GetIndexerState(ctx, pi.name)
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows {
			// Initialize with block 0
			if err := pi.InitializeState(ctx, 0); err != nil {
				return fmt.Errorf("failed to initialize state: %w", err)
			}
			state, _ = pi.queries.GetIndexerState(ctx, pi.name)
		} else {
			return fmt.Errorf("failed to get indexer state: %w", err)
		}
	}

	pi.logger.Info("Starting plays indexer",
		zap.Int64("lastIndexedBlock", state.LastIndexedBlock))

	// Main indexing loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := pi.indexOnce(ctx); err != nil {
				pi.logger.Error("Error during indexing", zap.Error(err))
				if setErr := pi.SetError(ctx, err); setErr != nil {
					pi.logger.Error("Failed to set error state", zap.Error(setErr))
				}
				// Sleep before retrying
				if err := pi.Sleep(ctx, 5*time.Second); err != nil {
					return err
				}
			}
		}
	}
}

func (pi *PlaysIndexer) indexOnce(ctx context.Context) error {
	// Get current state
	state, err := pi.queries.GetIndexerState(ctx, pi.name)
	if err != nil {
		return fmt.Errorf("failed to get indexer state: %w", err)
	}

	// Get the latest block height from blocks table
	latestHeight, err := pi.GetLatestBlockHeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block height: %w", err)
	}

	if latestHeight <= 0 {
		pi.logger.Debug("No blocks available yet")
		return pi.Sleep(ctx, pi.pollInterval)
	}

	// Update target block
	if err := pi.UpdateTarget(ctx, latestHeight); err != nil {
		pi.logger.Warn("Failed to update target block", zap.Error(err))
	}

	nextBlock := state.LastIndexedBlock + 1
	blocksRemaining := latestHeight - state.LastIndexedBlock

	if blocksRemaining <= 0 {
		pi.logger.Debug("At chain head", zap.Int64("height", state.LastIndexedBlock))
		return pi.Sleep(ctx, pi.pollInterval)
	}

	// Determine batch size and mode
	batchSize := pi.CalculateBatchSize(blocksRemaining)
	useBatch := pi.ShouldUseBatchMode(blocksRemaining)

	if useBatch {
		// Batch mode: process multiple blocks of transactions
		endBlock := nextBlock + int64(batchSize) - 1
		if endBlock > latestHeight {
			endBlock = latestHeight
		}

		pi.logger.Info("Batch indexing plays",
			zap.Int64("startBlock", nextBlock),
			zap.Int64("endBlock", endBlock),
			zap.Int32("batchSize", batchSize),
			zap.Int64("blocksRemaining", blocksRemaining))

		// Process batch
		if err := pi.processBatch(ctx, nextBlock, endBlock); err != nil {
			return fmt.Errorf("failed to process batch: %w", err)
		}

		// Update progress
		if err := pi.UpdateProgress(ctx, endBlock, "running"); err != nil {
			pi.logger.Warn("Failed to update progress", zap.Error(err))
		}

		// Short sleep in batch mode
		return pi.Sleep(ctx, pi.batchPollInterval)

	} else {
		// Single block mode
		pi.logger.Debug("Indexing plays from single block",
			zap.Int64("height", nextBlock),
			zap.Int64("blocksRemaining", blocksRemaining))

		// Process single block
		if err := pi.processBlock(ctx, nextBlock); err != nil {
			return fmt.Errorf("failed to process block %d: %w", nextBlock, err)
		}

		// Update progress
		if err := pi.UpdateProgress(ctx, nextBlock, "running"); err != nil {
			pi.logger.Warn("Failed to update progress", zap.Error(err))
		}

		// Normal sleep when near head
		return pi.Sleep(ctx, pi.pollInterval)
	}
}

func (pi *PlaysIndexer) processBatch(ctx context.Context, startBlock, endBlock int64) error {
	// Get transactions in the block range using sqlc
	transactions, err := pi.queries.GetTransactionsByTypeAndBlockRange(ctx, db.GetTransactionsByTypeAndBlockRangeParams{
		TxType:        "play",
		BlockHeight:   startBlock,
		BlockHeight_2: endBlock,
	})
	if err != nil {
		return fmt.Errorf("failed to query play transactions: %w", err)
	}

	// Start a transaction for batch insert
	tx, err := pi.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := pi.queries.WithTx(tx)
	playCount := 0

	for _, transaction := range transactions {
		// Parse transaction data
		if transaction.Data == nil {
			continue
		}

		var txProto corev1.Transaction
		if err := protojson.Unmarshal(transaction.Data, &txProto); err != nil {
			pi.logger.Warn("Failed to unmarshal transaction data",
				zap.String("txHash", transaction.TxHash),
				zap.Error(err))
			continue
		}

		// Extract plays from transaction
		if plays := txProto.Transaction.GetPlays(); plays != nil {
			for _, play := range plays.GetPlays() {
				err := qtx.InsertPlay(ctx, db.InsertPlayParams{
					UserID:      play.UserId,
					TrackID:     play.TrackId,
					City:        play.City,
					Region:      play.Region,
					Country:     play.Country,
					Latitude:    pgtype.Numeric{Valid: false}, // TODO: Extract from location if available
					Longitude:   pgtype.Numeric{Valid: false}, // TODO: Extract from location if available
					PlayedAt:    pgtype.Timestamp{Time: play.Timestamp.AsTime(), Valid: true},
					ListenedAt:  pgtype.Timestamp{Time: play.Timestamp.AsTime(), Valid: true},
					RecordedAt:  pgtype.Timestamp{Time: transaction.CreatedAt.Time, Valid: true},
					BlockHeight: transaction.BlockHeight,
					TxHash:      transaction.TxHash,
				})
				if err != nil {
					pi.logger.Warn("Failed to insert play",
						zap.String("txHash", transaction.TxHash),
						zap.Error(err))
				} else {
					playCount++
				}
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	pi.logger.Debug("Indexed plays batch",
		zap.Int64("startBlock", startBlock),
		zap.Int64("endBlock", endBlock),
		zap.Int("playCount", playCount))

	return nil
}

func (pi *PlaysIndexer) processBlock(ctx context.Context, blockHeight int64) error {
	// Get transactions for this specific block using sqlc
	transactions, err := pi.queries.GetTransactionsByTypeAndBlock(ctx, db.GetTransactionsByTypeAndBlockParams{
		TxType:      "play",
		BlockHeight: blockHeight,
	})
	if err != nil {
		return fmt.Errorf("failed to query play transactions: %w", err)
	}

	playCount := 0

	for _, transaction := range transactions {
		// Parse transaction data
		if transaction.Data == nil {
			continue
		}

		var txProto corev1.Transaction
		if err := protojson.Unmarshal(transaction.Data, &txProto); err != nil {
			pi.logger.Warn("Failed to unmarshal transaction data",
				zap.String("txHash", transaction.TxHash),
				zap.Error(err))
			continue
		}

		// Extract plays from transaction
		if plays := txProto.Transaction.GetPlays(); plays != nil {
			for _, play := range plays.GetPlays() {
				err := pi.queries.InsertPlay(ctx, db.InsertPlayParams{
					UserID:      play.UserId,
					TrackID:     play.TrackId,
					City:        play.City,
					Region:      play.Region,
					Country:     play.Country,
					Latitude:    pgtype.Numeric{Valid: false}, // TODO: Extract from location if available
					Longitude:   pgtype.Numeric{Valid: false}, // TODO: Extract from location if available
					PlayedAt:    pgtype.Timestamp{Time: play.Timestamp.AsTime(), Valid: true},
					ListenedAt:  pgtype.Timestamp{Time: play.Timestamp.AsTime(), Valid: true},
					RecordedAt:  pgtype.Timestamp{Time: transaction.CreatedAt.Time, Valid: true},
					BlockHeight: blockHeight,
					TxHash:      transaction.TxHash,
				})
				if err != nil {
					pi.logger.Warn("Failed to insert play",
						zap.String("txHash", transaction.TxHash),
						zap.Error(err))
				} else {
					playCount++
				}
			}
		}
	}

	if playCount > 0 {
		pi.logger.Debug("Indexed plays from block",
			zap.Int64("blockHeight", blockHeight),
			zap.Int("playCount", playCount))
	}

	return nil
}
