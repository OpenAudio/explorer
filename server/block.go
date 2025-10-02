package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/OpenAudio/explorer/db"
	"github.com/OpenAudio/explorer/templates/pages"
	"github.com/labstack/echo/v4"
)

func (s *Server) Blocks(c echo.Context) error {
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

	// Get blocks from database
	blocksData, err := s.db.GetBlocksByPage(c.Request().Context(), db.GetBlocksByPageParams{
		Limit:  count,
		Offset: offset,
	})
	if err != nil {
		s.logger.Warn("Failed to get blocks", "error", err)
		blocksData = []db.Block{}
	}

	// Convert to pointers
	blocks := make([]*db.Block, len(blocksData))
	blockTransactions := make([]int32, len(blocksData))
	for i := range blocksData {
		blocks[i] = &blocksData[i]
		// Get transaction count for each block
		txCount, err := s.db.GetBlockTransactionCount(c.Request().Context(), blocksData[i].Height)
		if err != nil {
			s.logger.Warn("Failed to get transaction count for block", "height", blocksData[i].Height, "error", err)
			txCount = 0
		}
		blockTransactions[i] = int32(txCount)
	}

	// Calculate pagination state
	hasNext := len(blocks) == int(count) // Simple check - if we got the full limit, there might be more
	hasPrev := page > 1

	props := pages.BlocksProps{
		Blocks:            blocks,
		BlockTransactions: blockTransactions,
		CurrentPage:       page,
		HasNext:           hasNext,
		HasPrev:           hasPrev,
		PageSize:          count,
	}

	p := pages.Blocks(props)
	ctx := c.Request().Context()
	return p.Render(ctx, c.Response().Writer)
}

func (s *Server) Block(c echo.Context) error {
	height, err := strconv.ParseInt(c.Param("height"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid block height")
	}

	ctx := c.Request().Context()

	// Get block by height
	block, err := s.db.GetBlockByHeight(ctx, height)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("Block not found at height %d", height))
	}

	// Get transactions for this block
	// First get all transactions and filter by block height
	// This is not the most efficient but will work for now - TODO: add GetTransactionsByBlockHeight query
	transactionsData, err := s.db.GetTransactionsByPage(ctx, db.GetTransactionsByPageParams{
		Limit:  1000, // Get a large number to ensure we get all for this block
		Offset: 0,
	})
	if err != nil {
		s.logger.Warn("Failed to get transactions", "error", err)
		transactionsData = []db.Transaction{}
	}

	// Filter transactions for this specific block height
	var blockTransactions []*db.Transaction
	for i := range transactionsData {
		if transactionsData[i].BlockHeight == height {
			blockTransactions = append(blockTransactions, &transactionsData[i])
		}
	}

	// Create block props
	props := pages.BlockProps{
		Block:        &block,
		Transactions: blockTransactions,
	}

	p := pages.Block(props)
	return p.Render(ctx, c.Response().Writer)
}
