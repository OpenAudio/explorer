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

// ValidatorsIndexer indexes validator-related transactions
type ValidatorsIndexer struct {
	*BaseIndexer
	pollInterval      time.Duration
	batchPollInterval time.Duration
}

// NewValidatorsIndexer creates a new validators indexer
func NewValidatorsIndexer(pool *pgxpool.Pool, logger *zap.Logger) *ValidatorsIndexer {
	return &ValidatorsIndexer{
		BaseIndexer:       NewBaseIndexer("validators", pool, logger, 100),
		pollInterval:      2 * time.Second,
		batchPollInterval: 200 * time.Millisecond,
	}
}

// Start begins the validators indexing process
func (vi *ValidatorsIndexer) Start(ctx context.Context) error {
	// Initialize state if needed
	state, err := vi.queries.GetIndexerState(ctx, vi.name)
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows {
			// Initialize with block 0
			if err := vi.InitializeState(ctx, 0); err != nil {
				return fmt.Errorf("failed to initialize state: %w", err)
			}
			state, _ = vi.queries.GetIndexerState(ctx, vi.name)
		} else {
			return fmt.Errorf("failed to get indexer state: %w", err)
		}
	}

	vi.logger.Info("Starting validators indexer",
		zap.Int64("lastIndexedBlock", state.LastIndexedBlock))

	// Main indexing loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := vi.indexOnce(ctx); err != nil {
				vi.logger.Error("Error during indexing", zap.Error(err))
				if setErr := vi.SetError(ctx, err); setErr != nil {
					vi.logger.Error("Failed to set error state", zap.Error(setErr))
				}
				// Sleep before retrying
				if err := vi.Sleep(ctx, 5*time.Second); err != nil {
					return err
				}
			}
		}
	}
}

func (vi *ValidatorsIndexer) indexOnce(ctx context.Context) error {
	// Get current state
	state, err := vi.queries.GetIndexerState(ctx, vi.name)
	if err != nil {
		return fmt.Errorf("failed to get indexer state: %w", err)
	}

	// Get the latest block height from blocks table
	latestHeight, err := vi.GetLatestBlockHeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block height: %w", err)
	}

	if latestHeight <= 0 {
		vi.logger.Debug("No blocks available yet")
		return vi.Sleep(ctx, vi.pollInterval)
	}

	// Update target block
	if err := vi.UpdateTarget(ctx, latestHeight); err != nil {
		vi.logger.Warn("Failed to update target block", zap.Error(err))
	}

	nextBlock := state.LastIndexedBlock + 1
	blocksRemaining := latestHeight - state.LastIndexedBlock

	if blocksRemaining <= 0 {
		vi.logger.Debug("At chain head", zap.Int64("height", state.LastIndexedBlock))
		return vi.Sleep(ctx, vi.pollInterval)
	}

	// Determine batch size and mode
	batchSize := vi.CalculateBatchSize(blocksRemaining)
	useBatch := vi.ShouldUseBatchMode(blocksRemaining)

	if useBatch {
		// Batch mode: process multiple blocks
		endBlock := nextBlock + int64(batchSize) - 1
		if endBlock > latestHeight {
			endBlock = latestHeight
		}

		vi.logger.Info("Batch indexing validators",
			zap.Int64("startBlock", nextBlock),
			zap.Int64("endBlock", endBlock),
			zap.Int32("batchSize", batchSize),
			zap.Int64("blocksRemaining", blocksRemaining))

		// Process batch
		if err := vi.processBatch(ctx, nextBlock, endBlock); err != nil {
			return fmt.Errorf("failed to process batch: %w", err)
		}

		// Update progress
		if err := vi.UpdateProgress(ctx, endBlock, "running"); err != nil {
			vi.logger.Warn("Failed to update progress", zap.Error(err))
		}

		// Short sleep in batch mode
		return vi.Sleep(ctx, vi.batchPollInterval)

	} else {
		// Single block mode
		vi.logger.Debug("Indexing validators from single block",
			zap.Int64("height", nextBlock),
			zap.Int64("blocksRemaining", blocksRemaining))

		// Process single block
		if err := vi.processBlock(ctx, nextBlock); err != nil {
			return fmt.Errorf("failed to process block %d: %w", nextBlock, err)
		}

		// Update progress
		if err := vi.UpdateProgress(ctx, nextBlock, "running"); err != nil {
			vi.logger.Warn("Failed to update progress", zap.Error(err))
		}

		// Normal sleep when near head
		return vi.Sleep(ctx, vi.pollInterval)
	}
}

func (vi *ValidatorsIndexer) processBatch(ctx context.Context, startBlock, endBlock int64) error {
	// Query validator-related transactions in the block range
	registrations, err := vi.queries.GetTransactionsByTypeAndBlockRange(ctx, db.GetTransactionsByTypeAndBlockRangeParams{
		TxType:        "validator_registration",
		BlockHeight:   startBlock,
		BlockHeight_2: endBlock,
	})
	if err != nil {
		return fmt.Errorf("failed to query registration transactions: %w", err)
	}

	deregistrations, err := vi.queries.GetTransactionsByTypeAndBlockRange(ctx, db.GetTransactionsByTypeAndBlockRangeParams{
		TxType:        "validator_deregistration",
		BlockHeight:   startBlock,
		BlockHeight_2: endBlock,
	})
	if err != nil {
		return fmt.Errorf("failed to query deregistration transactions: %w", err)
	}

	// Process registrations
	for _, tx := range registrations {
		if err := vi.processValidatorRegistration(ctx, tx); err != nil {
			vi.logger.Warn("Failed to process validator registration",
				zap.String("txHash", tx.TxHash),
				zap.Error(err))
		}
	}

	// Process deregistrations
	for _, tx := range deregistrations {
		if err := vi.processValidatorDeregistration(ctx, tx); err != nil {
			vi.logger.Warn("Failed to process validator deregistration",
				zap.String("txHash", tx.TxHash),
				zap.Error(err))
		}
	}

	return nil
}

func (vi *ValidatorsIndexer) processBlock(ctx context.Context, blockHeight int64) error {
	// Query validator-related transactions for this block
	registrations, err := vi.queries.GetTransactionsByTypeAndBlock(ctx, db.GetTransactionsByTypeAndBlockParams{
		TxType:      "validator_registration",
		BlockHeight: blockHeight,
	})
	if err != nil {
		return fmt.Errorf("failed to query registration transactions: %w", err)
	}

	deregistrations, err := vi.queries.GetTransactionsByTypeAndBlock(ctx, db.GetTransactionsByTypeAndBlockParams{
		TxType:      "validator_deregistration",
		BlockHeight: blockHeight,
	})
	if err != nil {
		return fmt.Errorf("failed to query deregistration transactions: %w", err)
	}

	// Process registrations
	for _, tx := range registrations {
		if err := vi.processValidatorRegistration(ctx, tx); err != nil {
			vi.logger.Warn("Failed to process validator registration",
				zap.String("txHash", tx.TxHash),
				zap.Error(err))
		}
	}

	// Process deregistrations
	for _, tx := range deregistrations {
		if err := vi.processValidatorDeregistration(ctx, tx); err != nil {
			vi.logger.Warn("Failed to process validator deregistration",
				zap.String("txHash", tx.TxHash),
				zap.Error(err))
		}
	}

	return nil
}

func (vi *ValidatorsIndexer) processValidatorRegistration(ctx context.Context, tx db.Transaction) error {
	if tx.Data == nil {
		return nil
	}

	// Parse transaction data
	var txProto corev1.Transaction
	if err := protojson.Unmarshal(tx.Data, &txProto); err != nil {
		return fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	// Extract validator registration
	attestation := txProto.Transaction.GetAttestation()
	if attestation == nil {
		return nil
	}

	vr := attestation.GetValidatorRegistration()
	if vr == nil {
		return nil
	}

	// Upsert validator
	err := vi.queries.UpsertValidator(ctx, db.UpsertValidatorParams{
		Address:      vr.DelegateWallet,
		CometAddress: vr.CometAddress,
		Endpoint:     vr.Endpoint,
		NodeType:     vr.NodeType,
		Spid:         vr.SpId,
		VotingPower:  vr.Power,
		Status:       "active",
		RegisteredAt: tx.BlockHeight,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert validator: %w", err)
	}

	// Insert validator event
	err = vi.queries.InsertValidatorEvent(ctx, db.InsertValidatorEventParams{
		ValidatorAddress: vr.DelegateWallet,
		EventType:        "registration",
		BlockHeight:      tx.BlockHeight,
		TxHash:           tx.TxHash,
	})
	if err != nil {
		vi.logger.Warn("Failed to insert validator event",
			zap.String("address", vr.DelegateWallet),
			zap.Error(err))
	}

	vi.logger.Debug("Processed validator registration",
		zap.String("address", vr.DelegateWallet),
		zap.String("txHash", tx.TxHash))

	return nil
}

func (vi *ValidatorsIndexer) processValidatorDeregistration(ctx context.Context, tx db.Transaction) error {
	if tx.Data == nil {
		return nil
	}

	// Parse transaction data
	var txProto corev1.Transaction
	if err := protojson.Unmarshal(tx.Data, &txProto); err != nil {
		return fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	// Extract validator deregistration
	attestation := txProto.Transaction.GetAttestation()
	if attestation == nil {
		return nil
	}

	vd := attestation.GetValidatorDeregistration()
	if vd == nil {
		return nil
	}

	// Look up validator by comet address
	validator, err := vi.queries.GetValidatorByCometAddress(ctx, vd.CometAddress)
	if err != nil {
		if err == sql.ErrNoRows {
			vi.logger.Warn("Validator not found for deregistration",
				zap.String("cometAddress", vd.CometAddress))
			return nil
		}
		return fmt.Errorf("failed to get validator: %w", err)
	}

	// Update validator status
	err = vi.queries.UpdateValidatorStatus(ctx, db.UpdateValidatorStatusParams{
		Address:        validator.Address,
		Status:         "deregistered",
		DeregisteredAt: pgtype.Int8{Int64: tx.BlockHeight, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update validator status: %w", err)
	}

	// Insert validator event
	err = vi.queries.InsertValidatorEvent(ctx, db.InsertValidatorEventParams{
		ValidatorAddress: validator.Address,
		EventType:        "deregistration",
		BlockHeight:      tx.BlockHeight,
		TxHash:           tx.TxHash,
	})
	if err != nil {
		vi.logger.Warn("Failed to insert validator event",
			zap.String("address", validator.Address),
			zap.Error(err))
	}

	vi.logger.Debug("Processed validator deregistration",
		zap.String("address", validator.Address),
		zap.String("txHash", tx.TxHash))

	return nil
}
