package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/OpenAudio/explorer/templates/pages"
	"github.com/OpenAudio/explorer/db"
	"github.com/labstack/echo/v4"
)

// getTransactionsWithBlockHeights is a helper method to get transactions with their block heights
func (s *Server) getTransactionsWithBlockHeights(ctx context.Context, limit, offset int32) ([]*db.Transaction, map[string]int64, error) {
	// Use GetTransactionsByPage for proper offset-based pagination
	transactions, err := s.db.GetTransactionsByPage(ctx, db.GetTransactionsByPageParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, nil, err
	}

	// Convert to pointers and create block heights map
	txPointers := make([]*db.Transaction, len(transactions))
	blockHeights := make(map[string]int64)
	for i := range transactions {
		txPointers[i] = &transactions[i]
		blockHeights[transactions[i].TxHash] = transactions[i].BlockHeight
	}

	return txPointers, blockHeights, nil
}

func (s *Server) Transactions(c echo.Context) error {
	// Parse query parameters
	pageParam := c.QueryParam("page")
	countParam := c.QueryParam("count")

	page := int32(1) // default to page 1
	if pageParam != "" {
		if parsedPage, err := strconv.ParseInt(pageParam, 10, 32); err == nil && parsedPage > 0 {
			page = int32(parsedPage)
		}
	}

	count := int32(50) // default to 50 per page
	if countParam != "" {
		if parsedCount, err := strconv.ParseInt(countParam, 10, 32); err == nil && parsedCount > 0 && parsedCount <= 200 {
			count = int32(parsedCount)
		}
	}

	// Calculate offset from page number
	offset := (page - 1) * count

	transactions, blockHeights, err := s.getTransactionsWithBlockHeights(c.Request().Context(), count, offset)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get transactions")
	}

	// Calculate pagination state
	hasNext := len(transactions) == int(count) // Simple check - if we got the full limit, there might be more
	hasPrev := page > 1

	props := pages.TransactionsProps{
		Transactions: transactions,
		BlockHeights: blockHeights,
		CurrentPage:  page,
		HasNext:      hasNext,
		HasPrev:      hasPrev,
		PageSize:     count,
	}

	p := pages.Transactions(props)
	ctx := c.Request().Context()
	return p.Render(ctx, c.Response().Writer)
}

func (s *Server) Transaction(c echo.Context) error {
	txHash := c.Param("hash")
	if txHash == "" {
		return c.String(http.StatusBadRequest, "Transaction hash required")
	}

	ctx := c.Request().Context()

	// Get transaction by hash
	transaction, err := s.db.GetTransactionByHash(ctx, txHash)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("Transaction not found: %s", txHash))
	}

	// Get block info for this transaction
	block, err := s.db.GetBlockByHeight(ctx, transaction.BlockHeight)
	if err != nil {
		s.logger.Warn("Failed to get block for transaction", "blockHeight", transaction.BlockHeight, "error", err)
		return c.String(http.StatusNotFound, fmt.Sprintf("Block not found at height %d", transaction.BlockHeight))
	}

	// Fetch transaction content based on type
	var content interface{}
	switch transaction.TxType {
	case "play":
		plays, err := s.db.GetPlaysByTxHash(ctx, txHash)
		if err != nil {
			s.logger.Warn("Failed to get plays for transaction", "txHash", txHash, "error", err)
		} else if len(plays) > 0 {
			// Convert to pointers for template
			playPointers := make([]*db.Play, len(plays))
			for i := range plays {
				playPointers[i] = &plays[i]
			}
			content = playPointers
		}

	case "manage_entity":
		entity, err := s.db.GetManageEntityByTxHash(ctx, txHash)
		if err != nil {
			s.logger.Warn("Failed to get manage entity for transaction", "txHash", txHash, "error", err)
		} else {
			content = &entity
		}

	case "validator_registration":
		registration, err := s.db.GetValidatorRegistrationByTxHash(ctx, txHash)
		if err != nil {
			s.logger.Warn("Failed to get validator registration for transaction", "txHash", txHash, "error", err)
		} else {
			content = &registration
		}

	case "validator_deregistration":
		deregistration, err := s.db.GetValidatorDeregistrationByTxHash(ctx, txHash)
		if err != nil {
			s.logger.Warn("Failed to get validator deregistration for transaction", "txHash", txHash, "error", err)
		} else {
			content = &deregistration
		}

	case "sla_rollup":
		slaRollup, err := s.db.GetSlaRollupByTxHash(ctx, txHash)
		if err != nil {
			s.logger.Warn("Failed to get SLA rollup for transaction", "txHash", txHash, "error", err)
		} else {
			content = &slaRollup
		}

	case "storage_proof":
		storageProof, err := s.db.GetStorageProofByTxHash(ctx, txHash)
		if err != nil {
			s.logger.Warn("Failed to get storage proof for transaction", "txHash", txHash, "error", err)
		} else {
			content = &storageProof
		}

	case "storage_proof_verification":
		storageProofVerification, err := s.db.GetStorageProofVerificationByTxHash(ctx, txHash)
		if err != nil {
			s.logger.Warn("Failed to get storage proof verification for transaction", "txHash", txHash, "error", err)
		} else {
			content = &storageProofVerification
		}
	}

	// Create transaction props
	props := pages.TransactionProps{
		Transaction: &transaction,
		Proposer:    block.ProposerAddress,
		Content:     content,
	}

	p := pages.Transaction(props)
	return p.Render(ctx, c.Response().Writer)
}
