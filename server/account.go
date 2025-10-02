package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/OpenAudio/explorer/db"
	"github.com/OpenAudio/explorer/templates/pages"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

func (s *Server) Account(c echo.Context) error {
	address := c.Param("address")
	if address == "" {
		return c.String(http.StatusBadRequest, "Address parameter is required")
	}

	isEthAddress := ethcommon.IsHexAddress(address)
	if !isEthAddress {
		// assume handle and query audius api
		res, err := http.Get(fmt.Sprintf("https://api.audius.co/v1/users/handle/%s", address))
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid address")
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid address")
		}

		type audiusUser struct {
			Wallet string `json:"wallet"`
		}

		type audiusResponse struct {
			Data audiusUser `json:"data"`
		}

		var response audiusResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid address")
		}
		address = response.Data.Wallet

		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/account/%s", address))
	}

	// Parse query parameters
	pageParam := c.QueryParam("page")
	countParam := c.QueryParam("count")
	relationFilter := c.QueryParam("relation")
	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

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

	ctx := c.Request().Context()
	etlDB := s.db

	// Parse date filters
	var startTimestamp, endTimestamp pgtype.Timestamp
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			startTimestamp = pgtype.Timestamp{Time: t, Valid: true}
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			// Add 24 hours to include the entire end date
			endTimestamp = pgtype.Timestamp{Time: t.Add(24 * time.Hour), Valid: true}
		}
	}

	// Get transactions for this address
	transactionRows, err := etlDB.GetTransactionsByAddress(ctx, db.GetTransactionsByAddressParams{
		Lower:   address,
		Column2: relationFilter, // empty string means all relations
		Column3: startTimestamp,
		Column4: endTimestamp,
		Limit:   count,
		Offset:  offset,
	})
	if err != nil {
		s.logger.Error("Failed to get transactions for address", "address", address, "error", err)
		return c.String(http.StatusInternalServerError, "Failed to get transactions")
	}

	// Get total count for pagination
	totalCount, err := etlDB.GetTransactionCountByAddress(ctx, db.GetTransactionCountByAddressParams{
		Lower:   address,
		Column2: relationFilter,
		Column3: startTimestamp,
		Column4: endTimestamp,
	})
	if err != nil {
		s.logger.Error("Failed to get transaction count for address", "address", address, "error", err)
		return c.String(http.StatusInternalServerError, "Failed to get transaction count")
	}

	// Get available relation types for filter dropdown
	relationTypesRaw, err := etlDB.GetRelationTypesByAddress(ctx, address)
	if err != nil {
		s.logger.Error("Failed to get relation types for address", "address", address, "error", err)
		// Don't fail the request, just log the error
		relationTypesRaw = []string{}
	}

	// Convert interface{} slice to string slice
	relationTypes := make([]string, len(relationTypesRaw))
	for i, rt := range relationTypesRaw {
		relationTypes[i] = rt
	}

	// Convert transaction rows to transactions and extract relations
	transactions := make([]*db.Transaction, len(transactionRows))
	txRelations := make([]string, len(transactionRows))
	for i, row := range transactionRows {
		transactions[i] = &db.Transaction{
			ID:          row.ID,
			TxHash:      row.TxHash,
			BlockHeight: row.BlockHeight,
			TxIndex:     row.TxIndex,
			TxType:      row.TxType,
			CreatedAt:   row.CreatedAt,
		}
		// Handle relation type assertion
		txRelations[i] = row.Relation
	}

	// Calculate pagination state
	hasNext := int64(offset+count) < totalCount
	hasPrev := page > 1

	props := pages.AccountProps{
		Address:       address,
		Transactions:  transactions,
		TxRelations:   txRelations,
		CurrentPage:   page,
		HasNext:       hasNext,
		HasPrev:       hasPrev,
		PageSize:      count,
		RelationTypes: relationTypes,
		CurrentFilter: relationFilter,
		StartDate:     startDate,
		EndDate:       endDate,
	}

	p := pages.Account(props)
	ctx = c.Request().Context()
	return p.Render(ctx, c.Response().Writer)
}
