package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/OpenAudio/explorer/db"
	"github.com/OpenAudio/explorer/templates/pages"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

func (s *Server) Validators(c echo.Context) error {
	// Parse query parameters
	pageParam := c.QueryParam("page")
	countParam := c.QueryParam("count")
	queryType := c.QueryParam("type") // "active", "registrations", "deregistrations"
	endpointFilter := c.QueryParam("endpoint_filter")

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

	// Default to active validators
	if queryType == "" {
		queryType = "active"
	}

	// Calculate offset from page number
	offset := (page - 1) * count

	var validators []*db.Validator
	validatorUptimeMap := make(map[string][]*db.SlaNodeReport)

	ctx := c.Request().Context()

	switch queryType {
	case "active":
		// Get active validators
		validatorsData, err := s.db.GetActiveValidators(ctx, db.GetActiveValidatorsParams{
			Limit:  count,
			Offset: offset,
		})
		if err != nil {
			s.logger.Warn("Failed to get active validators", "error", err)
			validatorsData = []db.Validator{}
		}

		// Convert to pointers and apply endpoint filter
		for i := range validatorsData {
			if endpointFilter == "" || strings.Contains(strings.ToLower(validatorsData[i].Endpoint), strings.ToLower(endpointFilter)) {
				validators = append(validators, &validatorsData[i])

				// Get uptime data for each validator
				reports, err := s.db.GetSlaNodeReportsByAddress(ctx, db.GetSlaNodeReportsByAddressParams{
					Lower: validatorsData[i].CometAddress,
					Limit: 5, // Get last 5 SLA reports
				})
				if err != nil {
					s.logger.Warn("Failed to get SLA reports", "address", validatorsData[i].CometAddress, "error", err)
				} else {
					reportPointers := make([]*db.SlaNodeReport, len(reports))
					for j := range reports {
						reportPointers[j] = &reports[j]
					}
					validatorUptimeMap[validatorsData[i].CometAddress] = reportPointers
				}
			}
		}

	case "registrations":
		// Get validator registrations - this will need a different approach since it's a different table
		regsData, err := s.db.GetValidatorRegistrations(ctx, db.GetValidatorRegistrationsParams{
			Limit:  count,
			Offset: offset,
		})
		if err != nil {
			s.logger.Warn("Failed to get validator registrations", "error", err)
			regsData = []db.GetValidatorRegistrationsRow{}
		}

		// Convert registrations to validator format for template
		for i := range regsData {
			validator := &db.Validator{
				Address:      regsData[i].Address,
				Endpoint:     regsData[i].Endpoint, // Already a string
				CometAddress: regsData[i].CometAddress,
				NodeType:     regsData[i].NodeType,    // Already a string
				Spid:         regsData[i].Spid,        // Already a string
				VotingPower:  regsData[i].VotingPower, // Already int64
				Status:       "registered",
				RegisteredAt: regsData[i].BlockHeight,
				CreatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true}, // Manual timestamp
			}
			if endpointFilter == "" || strings.Contains(strings.ToLower(validator.Endpoint), strings.ToLower(endpointFilter)) {
				validators = append(validators, validator)
			}
		}

	case "deregistrations":
		// Get validator deregistrations
		deregsData, err := s.db.GetValidatorDeregistrations(ctx, db.GetValidatorDeregistrationsParams{
			Limit:  count,
			Offset: offset,
		})
		if err != nil {
			s.logger.Warn("Failed to get validator deregistrations", "error", err)
			deregsData = []db.GetValidatorDeregistrationsRow{}
		}

		// Convert deregistrations to validator format for template
		for i := range deregsData {
			endpoint := ""
			if deregsData[i].Endpoint.Valid {
				endpoint = deregsData[i].Endpoint.String
			}
			nodeType := ""
			if deregsData[i].NodeType.Valid {
				nodeType = deregsData[i].NodeType.String
			}
			spid := ""
			if deregsData[i].Spid.Valid {
				spid = deregsData[i].Spid.String
			}
			votingPower := int64(0)
			if deregsData[i].VotingPower.Valid {
				votingPower = deregsData[i].VotingPower.Int64
			}

			validator := &db.Validator{
				Address:      "",
				Endpoint:     endpoint,
				CometAddress: deregsData[i].CometAddress,
				NodeType:     nodeType,
				Spid:         spid,
				VotingPower:  votingPower,
				Status:       "deregistered",
				RegisteredAt: deregsData[i].BlockHeight,
				CreatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true}, // placeholder
			}
			if endpointFilter == "" || strings.Contains(strings.ToLower(validator.Endpoint), strings.ToLower(endpointFilter)) {
				validators = append(validators, validator)
			}
		}
	}

	// Calculate pagination state
	hasNext := len(validators) == int(count) // Simple check - if we got the full limit, there might be more
	hasPrev := page > 1

	props := pages.ValidatorsProps{
		Validators:         validators,
		ValidatorUptimeMap: validatorUptimeMap,
		CurrentPage:        page,
		HasNext:            hasNext,
		HasPrev:            hasPrev,
		PageSize:           count,
		QueryType:          queryType,
		EndpointFilter:     endpointFilter,
	}

	p := pages.Validators(props)
	return p.Render(ctx, c.Response().Writer)
}

func (s *Server) Validator(c echo.Context) error {
	address := c.Param("address")
	if address == "" {
		return c.String(http.StatusBadRequest, "Validator address required")
	}

	ctx := c.Request().Context()

	// Get validator by address
	validator, err := s.db.GetValidatorByAddress(ctx, address)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("Validator not found: %s", address))
	}

	// Get SLA rollup reports for this validator
	reports, err := s.db.GetSlaNodeReportsByAddress(ctx, db.GetSlaNodeReportsByAddressParams{
		Lower: validator.CometAddress,
		Limit: 10, // Get last 10 reports
	})
	if err != nil {
		s.logger.Warn("Failed to get SLA reports for validator", "address", address, "error", err)
		reports = []db.SlaNodeReport{}
	}

	// Convert reports to pointers
	rollups := make([]*db.SlaNodeReport, len(reports))
	for i := range reports {
		rollups[i] = &reports[i]
	}

	// TODO: Get validator events from registration/deregistration tables
	// For now, create empty events slice
	events := []*pages.ValidatorEvent{}

	props := pages.ValidatorProps{
		Validator: &validator,
		Events:    events,
		Rollups:   rollups,
	}

	p := pages.Validator(props)
	return p.Render(ctx, c.Response().Writer)
}

func (s *Server) ValidatorsUptime(c echo.Context) error {
	// Parse query parameters for pagination
	pageParam := c.QueryParam("page")
	countParam := c.QueryParam("count")

	page := int32(1) // default to page 1
	if pageParam != "" {
		if parsedPage, err := strconv.ParseInt(pageParam, 10, 32); err == nil && parsedPage > 0 {
			page = int32(parsedPage)
		}
	}

	count := int32(20) // default to 20 per page for rollups
	if countParam != "" {
		if parsedCount, err := strconv.ParseInt(countParam, 10, 32); err == nil && parsedCount > 0 && parsedCount <= 100 {
			count = int32(parsedCount)
		}
	}

	// Calculate offset from page number
	offset := (page - 1) * count

	ctx := c.Request().Context()

	// Get paginated SLA rollups
	rollupsData, err := s.db.GetSlaRollupsWithPagination(ctx, db.GetSlaRollupsWithPaginationParams{
		Limit:  count,
		Offset: offset,
	})
	if err != nil {
		s.logger.Warn("Failed to get SLA rollups", "error", err)
		rollupsData = []db.SlaRollup{}
	}

	// Convert to pointers
	rollups := make([]*db.SlaRollup, len(rollupsData))
	for i := range rollupsData {
		rollups[i] = &rollupsData[i]
	}

	// Calculate pagination state
	hasNext := len(rollupsData) == int(count)
	hasPrev := page > 1

	// TODO: Get actual total count from database
	totalCount := int64(len(rollupsData)) // Placeholder

	props := pages.RollupsProps{
		Rollups:          rollups,
		RollupValidators: []*db.SlaNodeReport{}, // Not needed for rollups list view
		CurrentPage:      page,
		HasNext:          hasNext,
		HasPrev:          hasPrev,
		PageSize:         count,
		TotalCount:       totalCount,
	}

	p := pages.Rollups(props)
	return p.Render(ctx, c.Response().Writer)
}

func (s *Server) ValidatorsUptimeByRollup(c echo.Context) error {
	rollupIDParam := c.Param("rollupid")
	if rollupIDParam == "" {
		return c.String(http.StatusBadRequest, "Rollup ID required")
	}

	rollupID, err := strconv.ParseInt(rollupIDParam, 10, 32)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid rollup ID")
	}

	ctx := c.Request().Context()

	// First, get the actual SLA rollup data to get tx_hash, created_at, block quota, etc.
	rollupInfo, err := s.db.GetSlaRollupById(ctx, int32(rollupID))
	if err != nil {
		s.logger.Warn("Failed to get SLA rollup by ID", "rollupID", rollupID, "error", err)
		return c.String(http.StatusNotFound, fmt.Sprintf("SLA rollup not found: %d", rollupID))
	}

	// Get validators for this specific SLA rollup
	validatorsData, err := s.db.GetValidatorsForSlaRollup(ctx, int32(rollupID))
	if err != nil {
		s.logger.Warn("Failed to get validators for SLA rollup", "rollupID", rollupID, "error", err)
		validatorsData = []db.GetValidatorsForSlaRollupRow{}
	}

	// Calculate challenge statistics dynamically for this rollup's block range
	// This ensures we get the current accurate data instead of potentially stale pre-calculated values
	challengeStats, err := s.db.GetChallengeStatisticsForBlockRange(ctx, db.GetChallengeStatisticsForBlockRangeParams{
		BlockStart: rollupInfo.BlockStart,
		BlockEnd:   rollupInfo.BlockEnd,
	})
	if err != nil {
		s.logger.Warn("Failed to get challenge statistics", "rollupID", rollupID, "error", err)
		challengeStats = []db.GetChallengeStatisticsForBlockRangeRow{}
	}

	// Create a map for quick lookup of challenge statistics by address
	challengeStatsMap := make(map[string]db.GetChallengeStatisticsForBlockRangeRow)
	for _, stat := range challengeStats {
		challengeStatsMap[stat.Address] = stat
	}

	// Build validator uptime info for each validator
	validators := make([]*pages.ValidatorUptimeInfo, 0, len(validatorsData))
	for i := range validatorsData {
		validator := &db.Validator{
			Address:      validatorsData[i].Address,
			Endpoint:     validatorsData[i].Endpoint,
			CometAddress: validatorsData[i].CometAddress,
			NodeType:     validatorsData[i].NodeType,
			Spid:         validatorsData[i].Spid,
			VotingPower:  validatorsData[i].VotingPower,
			Status:       validatorsData[i].Status,
			RegisteredAt: validatorsData[i].RegisteredAt,
			CreatedAt:    validatorsData[i].CreatedAt,
			UpdatedAt:    validatorsData[i].UpdatedAt,
		}

		// Create a full SLA report for this rollup with all the required fields
		var reportPointers []*db.SlaNodeReport
		slaReport := &db.SlaNodeReport{
			SlaRollupID:        int32(rollupID),
			ValidatorAddress:   validatorsData[i].CometAddress,
			NumBlocksProposed:  0, // Default to 0
			ChallengesReceived: 0, // Default to 0
			ChallengesFailed:   0, // Default to 0
			CreatedAt:          rollupInfo.CreatedAt,
		}

		// Override with actual data if validator has report data (for blocks proposed)
		if validatorsData[i].NumBlocksProposed.Valid {
			slaReport.NumBlocksProposed = validatorsData[i].NumBlocksProposed.Int32
		}

		// Use dynamically calculated challenge statistics instead of potentially stale pre-calculated values
		if stat, exists := challengeStatsMap[validatorsData[i].CometAddress]; exists {
			slaReport.ChallengesReceived = int32(stat.ChallengesReceived)
			slaReport.ChallengesFailed = int32(stat.ChallengesFailed)
		}

		reportPointers = []*db.SlaNodeReport{slaReport}

		validators = append(validators, &pages.ValidatorUptimeInfo{
			Validator:     validator,
			RecentRollups: reportPointers,
		})
	}

	props := pages.ValidatorsUptimeByRollupProps{
		Validators: validators,
		RollupID:   int32(rollupID),
		RollupData: &rollupInfo,
	}

	p := pages.ValidatorsUptimeByRollup(props)
	return p.Render(ctx, c.Response().Writer)
}
