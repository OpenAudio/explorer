package indexers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/OpenAudio/explorer/db"
	corev1 "github.com/OpenAudio/go-openaudio/pkg/api/core/v1"
	"github.com/OpenAudio/go-openaudio/pkg/sdk"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

// BlocksIndexer indexes blocks from the RPC to Postgres
type BlocksIndexer struct {
	*BaseIndexer
	sdk               *sdk.OpenAudioSDK
	startingBlock     int64
	checkReadiness    bool
	pollInterval      time.Duration
	batchPollInterval time.Duration
}

// NewBlocksIndexer creates a new blocks indexer
func NewBlocksIndexer(
	pool *pgxpool.Pool,
	sdk *sdk.OpenAudioSDK,
	logger *zap.Logger,
	startingBlock int64,
	checkReadiness bool,
) *BlocksIndexer {
	return &BlocksIndexer{
		BaseIndexer:       NewBaseIndexer("blocks", pool, logger, 100),
		sdk:               sdk,
		startingBlock:     startingBlock,
		checkReadiness:    checkReadiness,
		pollInterval:      1 * time.Second,
		batchPollInterval: 100 * time.Millisecond,
	}
}

// Start begins the blocks indexing process
func (bi *BlocksIndexer) Start(ctx context.Context) error {
	// Wait for readiness if required
	if bi.checkReadiness {
		if err := bi.awaitReadiness(ctx); err != nil {
			return fmt.Errorf("failed to await readiness: %w", err)
		}
	}

	// Initialize state if needed
	state, err := bi.queries.GetIndexerState(ctx, bi.name)
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows {
			// Initialize with starting block
			startBlock := bi.startingBlock
			if startBlock <= 0 {
				startBlock = 1
			}
			if err := bi.InitializeState(ctx, startBlock-1); err != nil {
				return fmt.Errorf("failed to initialize state: %w", err)
			}
			state, _ = bi.queries.GetIndexerState(ctx, bi.name)
		} else {
			return fmt.Errorf("failed to get indexer state: %w", err)
		}
	}

	bi.logger.Info("Starting blocks indexer",
		zap.Int64("lastIndexedBlock", state.LastIndexedBlock),
		zap.Int64("startingBlock", bi.startingBlock))

	// Main indexing loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := bi.indexOnce(ctx); err != nil {
				bi.logger.Error("Error during indexing", zap.Error(err))
				if setErr := bi.SetError(ctx, err); setErr != nil {
					bi.logger.Error("Failed to set error state", zap.Error(setErr))
				}
				// Sleep before retrying
				if err := bi.Sleep(ctx, 5*time.Second); err != nil {
					return err
				}
			}
		}
	}
}

func (bi *BlocksIndexer) indexOnce(ctx context.Context) error {
	// Get current state
	state, err := bi.queries.GetIndexerState(ctx, bi.name)
	if err != nil {
		return fmt.Errorf("failed to get indexer state: %w", err)
	}

	// Get the latest block height from the RPC
	statusResp, err := bi.sdk.Core.GetStatus(ctx, connect.NewRequest(&corev1.GetStatusRequest{}))
	if err != nil {
		return fmt.Errorf("failed to get chain status: %w", err)
	}

	latestHeight := statusResp.Msg.ChainInfo.CurrentHeight
	if latestHeight <= 0 {
		bi.logger.Debug("No blocks available yet")
		return bi.Sleep(ctx, bi.pollInterval)
	}

	// Update target block
	if err := bi.UpdateTarget(ctx, latestHeight); err != nil {
		bi.logger.Warn("Failed to update target block", zap.Error(err))
	}

	nextBlock := state.LastIndexedBlock + 1
	blocksRemaining := latestHeight - state.LastIndexedBlock

	if blocksRemaining <= 0 {
		bi.logger.Debug("At chain head", zap.Int64("height", state.LastIndexedBlock))
		return bi.Sleep(ctx, bi.pollInterval)
	}

	// Determine batch size and mode
	batchSize := bi.CalculateBatchSize(blocksRemaining)
	useBatch := bi.ShouldUseBatchMode(blocksRemaining)

	if useBatch {
		// Batch mode: fetch and index multiple blocks
		endBlock := nextBlock + int64(batchSize) - 1
		if endBlock > latestHeight {
			endBlock = latestHeight
		}

		bi.logger.Info("Batch indexing blocks",
			zap.Int64("startBlock", nextBlock),
			zap.Int64("endBlock", endBlock),
			zap.Int32("batchSize", batchSize),
			zap.Int64("blocksRemaining", blocksRemaining))

		// Build array of heights to fetch
		heights := make([]int64, 0, endBlock-nextBlock+1)
		for h := nextBlock; h <= endBlock; h++ {
			heights = append(heights, h)
		}

		// Fetch blocks batch
		blocksResp, err := bi.sdk.Core.GetBlocks(ctx, connect.NewRequest(&corev1.GetBlocksRequest{
			Height: heights,
		}))
		if err != nil {
			return fmt.Errorf("failed to get blocks batch: %w", err)
		}

		// Convert map to slice
		blocks := make([]*corev1.Block, 0, len(blocksResp.Msg.Blocks))
		for _, block := range blocksResp.Msg.Blocks {
			blocks = append(blocks, block)
		}

		// Index the batch
		if err := bi.indexBlocksBatch(ctx, blocks); err != nil {
			return fmt.Errorf("failed to index blocks batch: %w", err)
		}

		// Update progress
		if err := bi.UpdateProgress(ctx, endBlock, "running"); err != nil {
			bi.logger.Warn("Failed to update progress", zap.Error(err))
		}

		// Short sleep in batch mode
		return bi.Sleep(ctx, bi.batchPollInterval)

	} else {
		// Single block mode
		bi.logger.Debug("Indexing single block",
			zap.Int64("height", nextBlock),
			zap.Int64("blocksRemaining", blocksRemaining))

		blockResp, err := bi.sdk.Core.GetBlocks(ctx, connect.NewRequest(&corev1.GetBlocksRequest{
			Height: []int64{nextBlock},
		}))
		if err != nil {
			return fmt.Errorf("failed to get block %d: %w", nextBlock, err)
		}

		// Index single block
		// GetBlock returns a single block at the height
		if err := bi.indexBlock(ctx, blockResp.Msg.Blocks[nextBlock]); err != nil {
			return fmt.Errorf("failed to index block %d: %w", nextBlock, err)
		}

		// Update progress
		if err := bi.UpdateProgress(ctx, nextBlock, "running"); err != nil {
			bi.logger.Warn("Failed to update progress", zap.Error(err))
		}

		// Normal sleep when near head
		return bi.Sleep(ctx, bi.pollInterval)
	}
}

func (bi *BlocksIndexer) indexBlocksBatch(ctx context.Context, blocks []*corev1.Block) error {
	// Start a transaction for batch insert
	tx, err := bi.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := bi.queries.WithTx(tx)

	for _, block := range blocks {
		if err := bi.insertBlockAndTransactions(ctx, qtx, block); err != nil {
			return fmt.Errorf("failed to insert block %d: %w", block.Height, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	bi.logger.Debug("Indexed blocks batch",
		zap.Int("count", len(blocks)),
		zap.Int64("firstHeight", blocks[0].Height),
		zap.Int64("lastHeight", blocks[len(blocks)-1].Height))

	return nil
}

func (bi *BlocksIndexer) indexBlock(ctx context.Context, block *corev1.Block) error {
	return bi.insertBlockAndTransactions(ctx, bi.queries, block)
}

func (bi *BlocksIndexer) insertBlockAndTransactions(ctx context.Context, q *db.Queries, block *corev1.Block) error {
	// Marshal block data to JSON
	blockData, err := protojson.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block data: %w", err)
	}

	// Insert or update the block
	err = q.UpsertBlock(ctx, db.UpsertBlockParams{
		Height:          block.Height,
		Hash:            block.Hash,
		BlockTime:       pgtype.Timestamp{Time: block.Timestamp.AsTime(), Valid: true},
		ProposerAddress: block.Proposer,
		Data:            blockData,
	})
	if err != nil {
		return fmt.Errorf("failed to insert block: %w", err)
	}

	// Insert transactions
	for idx, tx := range block.Transactions {
		txType := bi.getTransactionType(tx.Transaction)
		sender := bi.getTransactionSender(tx.Transaction)

		// Marshal transaction data
		txData, err := protojson.Marshal(tx)
		if err != nil {
			bi.logger.Warn("Failed to marshal transaction data",
				zap.String("txHash", tx.Hash),
				zap.Error(err))
			txData = nil
		}

		err = q.UpsertTransaction(ctx, db.UpsertTransactionParams{
			TxHash:      tx.Hash,
			BlockHeight: block.Height,
			TxIndex:     int32(idx),
			TxType:      txType,
			Proposer:    block.Proposer,
			Sender:      sender,
			Data:        txData,
		})
		if err != nil {
			bi.logger.Warn("Failed to insert transaction",
				zap.String("txHash", tx.Hash),
				zap.Error(err))
		}
	}

	return nil
}

func (bi *BlocksIndexer) getTransactionType(tx *corev1.SignedTransaction) string {
	if tx == nil || tx.Transaction == nil {
		return "unknown"
	}

	switch tx.Transaction.(type) {
	case *corev1.SignedTransaction_Plays:
		return "play"
	case *corev1.SignedTransaction_ManageEntity:
		return "manage_entity"
	case *corev1.SignedTransaction_ValidatorRegistration:
		return "validator_registration_legacy"
	case *corev1.SignedTransaction_ValidatorDeregistration:
		return "validator_misbehavior_deregistration"
	case *corev1.SignedTransaction_SlaRollup:
		return "sla_rollup"
	case *corev1.SignedTransaction_StorageProof:
		return "storage_proof"
	case *corev1.SignedTransaction_StorageProofVerification:
		return "storage_proof_verification"
	case *corev1.SignedTransaction_Attestation:
		at := tx.GetAttestation()
		if at.GetValidatorRegistration() != nil {
			return "validator_registration"
		}
		if at.GetValidatorDeregistration() != nil {
			return "validator_deregistration"
		}
		return "attestation"
	default:
		return "unknown"
	}
}

func (bi *BlocksIndexer) getTransactionSender(tx *corev1.SignedTransaction) string {
	if tx == nil || tx.Transaction == nil {
		return ""
	}

	switch signedTx := tx.Transaction.(type) {
	case *corev1.SignedTransaction_Plays:
		if plays := signedTx.Plays.GetPlays(); len(plays) > 0 {
			return plays[0].UserId
		}
	case *corev1.SignedTransaction_ManageEntity:
		return signedTx.ManageEntity.GetSigner()
	case *corev1.SignedTransaction_StorageProof:
		return signedTx.StorageProof.GetAddress()
	case *corev1.SignedTransaction_Attestation:
		if vr := signedTx.Attestation.GetValidatorRegistration(); vr != nil {
			return vr.DelegateWallet
		}
		if vd := signedTx.Attestation.GetValidatorDeregistration(); vd != nil {
			return vd.CometAddress
		}
	case *corev1.SignedTransaction_ValidatorDeregistration:
		return signedTx.ValidatorDeregistration.CometAddress
	}

	return ""
}

func (bi *BlocksIndexer) awaitReadiness(ctx context.Context) error {
	bi.logger.Info("Awaiting chain readiness")
	attempts := 0
	maxAttempts := 60

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			attempts++
			if attempts > maxAttempts {
				return fmt.Errorf("timed out waiting for readiness after %d attempts", maxAttempts)
			}

			res, err := bi.sdk.Core.GetStatus(ctx, connect.NewRequest(&corev1.GetStatusRequest{}))
			if err != nil {
				bi.logger.Debug("Waiting for chain to be ready", zap.Int("attempt", attempts))
				if err := bi.Sleep(ctx, 1*time.Second); err != nil {
					return err
				}
				continue
			}

			// Check if chain has blocks (ready)
			if res.Msg.ChainInfo.CurrentHeight > 0 {
				bi.logger.Info("Chain is ready")
				return nil
			}

			if err := bi.Sleep(ctx, 1*time.Second); err != nil {
				return err
			}
		}
	}
}
