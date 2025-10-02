package server

import (
	"github.com/AudiusProject/audiusd/pkg/console/templates/pages"
	"github.com/AudiusProject/audiusd/pkg/etl/db"
	"github.com/labstack/echo/v4"
)

func (s *Server) Dashboard(c echo.Context) error {
	ctx := c.Request().Context()

	// Get dashboard transaction stats from materialized view
	txStats, err := s.etl.GetDB().GetDashboardTransactionStats(ctx)
	if err != nil {
		s.logger.Warn("Failed to get dashboard transaction stats", "error", err)
		// Use fallback empty stats
		txStats = db.MvDashboardTransactionStat{}
	}

	// Get transaction type breakdown from materialized view
	txTypes, err2 := s.etl.GetDB().GetDashboardTransactionTypes(ctx)
	if err2 != nil {
		s.logger.Warn("Failed to get dashboard transaction types", "error", err2)
		txTypes = []db.MvDashboardTransactionType{}
	}

	// Get latest indexed block
	latestBlockHeight, err := s.etl.GetDB().GetLatestIndexedBlock(ctx)
	if err != nil {
		s.logger.Warn("Failed to get latest block height", "error", err)
		latestBlockHeight = 0
	}

	// Get latest trusted block height from the trusted node
	trustedBlockHeight := int64(s.latestTrustedBlock.Load())

	// Calculate sync status - consider synced if within 60 blocks of the head
	const syncThreshold = 60
	var isSyncing bool
	var blockDelta int64
	syncProgressPercentage := float64(100) // Default to fully synced

	s.logger.Info("Latest block height", "height", latestBlockHeight)
	s.logger.Info("Trusted block height", "height", trustedBlockHeight)

	if trustedBlockHeight >= 0 && latestBlockHeight >= 0 {

		blockDelta = trustedBlockHeight - latestBlockHeight
		isSyncing = blockDelta > syncThreshold

		s.logger.Info("Block delta", "delta", blockDelta)
		s.logger.Info("Is syncing", "isSyncing", isSyncing)

		if isSyncing && trustedBlockHeight > 0 {
			syncProgressPercentage = (float64(latestBlockHeight) / float64(trustedBlockHeight)) * 100
			s.logger.Info("Sync progress percentage", "percentage", syncProgressPercentage)
			// Ensure percentage doesn't exceed 100
			if syncProgressPercentage > 100 {
				syncProgressPercentage = 100
			}
		} else {
			syncProgressPercentage = 100 // Fully synced
		}
	} else {
		s.logger.Info("No trusted block info, assuming synced")
		// If we don't have trusted block info, assume synced
		isSyncing = false
		blockDelta = 0
		syncProgressPercentage = 100
	}

	// Get latest SLA rollup for BPS/TPS data
	var bps, tps float64 = 0, 0
	var avgBlockTime float32 = 0
	latestSlaRollup, err := s.etl.GetDB().GetLatestSlaRollup(ctx)
	if err != nil {
		s.logger.Debug("Failed to get latest SLA rollup", "error", err)
		// Fall back to default values
		bps = 0.5
		tps = 0.1
		avgBlockTime = 2.0
	} else {
		bps = latestSlaRollup.Bps
		tps = latestSlaRollup.Tps
		// Calculate average block time from BPS (if BPS > 0)
		if bps > 0 {
			avgBlockTime = float32(1.0 / bps)
		} else {
			avgBlockTime = 2.0 // Default 2 seconds
		}
	}

	// Get some recent transactions for the dashboard
	transactions, blockHeights, err := s.getTransactionsWithBlockHeights(ctx, 10, 0)
	if err != nil {
		s.logger.Warn("Failed to get transactions", "error", err)
		transactions = []*db.EtlTransaction{}
		blockHeights = make(map[string]int64)
	}

	blocks, err := s.etl.GetDB().GetBlocksByPage(ctx, db.GetBlocksByPageParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		s.logger.Warn("Failed to get blocks", "error", err)
		return c.String(500, "Failed to get blocks")
	}

	blockPointers := make([]*db.EtlBlock, len(blocks))
	for i := range blocks {
		blockPointers[i] = &blocks[i]
	}

	// Get active validator count
	validatorCount, err := s.etl.GetDB().GetActiveValidatorCount(ctx)
	if err != nil {
		s.logger.Warn("Failed to get validator count", "error", err)
		validatorCount = 0
	}

	// Build stats using materialized view data
	stats := &pages.DashboardStats{
		CurrentBlockHeight:           latestBlockHeight,
		ChainID:                      s.etl.ChainID,
		BPS:                          bps,
		TPS:                          tps,
		TotalTransactions:            txStats.TotalTransactions,
		ValidatorCount:               validatorCount,
		LatestBlock:                  nil, // TODO: Implement
		RecentProposers:              nil, // TODO: Implement
		IsSyncing:                    isSyncing,
		LatestIndexedHeight:          latestBlockHeight,
		LatestChainHeight:            trustedBlockHeight,
		BlockDelta:                   blockDelta,
		TotalTransactions24h:         txStats.Transactions24h,
		TotalTransactionsPrevious24h: txStats.TransactionsPrevious24h,
		TotalTransactions7d:          txStats.Transactions7d,
		TotalTransactions30d:         txStats.Transactions30d,
		AvgBlockTime:                 avgBlockTime,
	}

	// Convert materialized view transaction types to template format
	maxTypes := 5 // only show up to 5 transaction types
	if len(txTypes) < maxTypes {
		maxTypes = len(txTypes)
	}
	transactionBreakdown := make([]*pages.TransactionTypeBreakdown, maxTypes)
	colors := []string{"bg-blue-500", "bg-green-500", "bg-purple-500", "bg-yellow-500", "bg-red-500", "bg-indigo-500", "bg-pink-500"}
	for i := 0; i < maxTypes; i++ {
		txType := txTypes[i]
		color := colors[i%len(colors)]
		transactionBreakdown[i] = &pages.TransactionTypeBreakdown{
			Type:  txType.TxType,
			Count: txType.TransactionCount,
			Color: color,
		}
	}

	// Get SLA performance data for the chart (most recent 50 rollups)
	slaRollupsData, err := s.etl.GetDB().GetSlaRollupsWithPagination(ctx, db.GetSlaRollupsWithPaginationParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		s.logger.Warn("Failed to get SLA rollups for performance chart", "error", err)
		slaRollupsData = []db.EtlSlaRollup{}
	}

	s.logger.Info("SLA rollups data retrieved", "count", len(slaRollupsData))

	// Build SLA performance data points for chart - Initialize as empty slice, not nil
	slaPerformanceData := make([]*pages.SLAPerformanceDataPoint, 0)

	// Build chart data if we have any rollups
	if len(slaRollupsData) > 0 {
		s.logger.Info("Building SLA performance chart data", "rollupCount", len(slaRollupsData))

		// Extract rollup IDs for healthy validators query
		rollupIDs := make([]int32, len(slaRollupsData))
		for i, rollup := range slaRollupsData {
			rollupIDs[i] = rollup.ID
		}

		// Get healthy validator counts for these rollups
		healthyValidatorData, err := s.etl.GetDB().GetHealthyValidatorCountsForRollups(ctx, rollupIDs)
		if err != nil {
			s.logger.Warn("Failed to get healthy validator counts", "error", err)
			healthyValidatorData = []db.GetHealthyValidatorCountsForRollupsRow{}
		}

		// Build a map for quick lookup of healthy validator counts
		healthyValidatorsMap := make(map[int32]int32)
		for _, hvData := range healthyValidatorData {
			if healthyCount, ok := hvData.HealthyValidators.(int64); ok {
				healthyValidatorsMap[hvData.RollupID] = int32(healthyCount)
			} else {
				healthyValidatorsMap[hvData.RollupID] = 0
			}
		}

		// Filter out invalid rollups and build valid data points
		validDataPoints := make([]*pages.SLAPerformanceDataPoint, 0, len(slaRollupsData))
		for i, rollup := range slaRollupsData {
			// Log the first rollup to see what data we're getting
			if i == 0 {
				s.logger.Info("First rollup data sample",
					"id", rollup.ID,
					"blockHeight", rollup.BlockHeight,
					"validatorCount", rollup.ValidatorCount,
					"bps", rollup.Bps,
					"tps", rollup.Tps,
					"createdAtValid", rollup.CreatedAt.Valid,
					"blockStart", rollup.BlockStart,
					"blockEnd", rollup.BlockEnd)
			}

			// Comprehensive validation of rollup data
			if rollup.ID <= 0 {
				s.logger.Debug("Skipping rollup with invalid ID", "index", i, "rollupId", rollup.ID)
				continue
			}

			if !rollup.CreatedAt.Valid {
				s.logger.Debug("Skipping rollup with invalid timestamp", "rollupId", rollup.ID)
				continue
			}

			if rollup.BlockHeight <= 0 {
				s.logger.Debug("Skipping rollup with invalid block height", "rollupId", rollup.ID, "blockHeight", rollup.BlockHeight)
				continue
			}

			if rollup.Bps < 0 || rollup.Tps < 0 {
				s.logger.Debug("Skipping rollup with invalid performance data", "rollupId", rollup.ID, "bps", rollup.Bps, "tps", rollup.Tps)
				continue
			}

			if rollup.BlockStart < 0 || rollup.BlockEnd <= 0 || rollup.BlockStart > rollup.BlockEnd {
				s.logger.Debug("Skipping rollup with invalid block range", "rollupId", rollup.ID, "start", rollup.BlockStart, "end", rollup.BlockEnd)
				continue
			}

			// Use the validator count from the rollup data itself
			validatorCount := rollup.ValidatorCount
			if validatorCount <= 0 {
				s.logger.Debug("Invalid validator count in rollup, using fallback", "rollupId", rollup.ID, "count", validatorCount)
				validatorCount = 1 // Minimum of 1 validator
			}

			// Get healthy validators count for this rollup
			healthyValidators := int32(0)
			if healthyCount, exists := healthyValidatorsMap[rollup.ID]; exists {
				healthyValidators = healthyCount
			}

			// Create a fully validated data point
			dataPoint := &pages.SLAPerformanceDataPoint{
				RollupID:          rollup.ID,
				BlockHeight:       rollup.BlockHeight,
				Timestamp:         rollup.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
				ValidatorCount:    validatorCount,
				HealthyValidators: healthyValidators,
				BPS:               rollup.Bps,
				TPS:               rollup.Tps,
				BlockStart:        rollup.BlockStart,
				BlockEnd:          rollup.BlockEnd,
			}

			// Extra safety check - ensure we're not adding nil
			if dataPoint != nil {
				validDataPoints = append(validDataPoints, dataPoint)
			}
		}

		// Use the data if we have any valid points after filtering
		if len(validDataPoints) > 0 {
			slaPerformanceData = validDataPoints
			s.logger.Info("Successfully built SLA performance data", "validPoints", len(validDataPoints))
		} else {
			s.logger.Warn("No valid rollup data points after filtering", "valid", len(validDataPoints), "total", len(slaRollupsData))
			// Keep the empty slice - don't set to nil
		}
	} else {
		s.logger.Info("No rollups available for chart", "rollupCount", len(slaRollupsData))
	}

	// Final debug of what we're passing to template
	s.logger.Info("Final SLA performance data for template", "dataPoints", len(slaPerformanceData))

	// Convert rollups to pointers for template
	recentSLARollups := make([]*db.EtlSlaRollup, len(slaRollupsData))
	for i := range slaRollupsData {
		recentSLARollups[i] = &slaRollupsData[i]
	}

	props := pages.DashboardProps{
		Stats:                  stats,
		TransactionBreakdown:   transactionBreakdown,
		RecentBlocks:           blockPointers,
		RecentTransactions:     transactions,
		RecentSLARollups:       recentSLARollups,
		SLAPerformanceData:     slaPerformanceData,
		BlockHeights:           blockHeights,
		SyncProgressPercentage: syncProgressPercentage,
	}

	p := pages.Dashboard(props)

	// Use context with environment
	return p.Render(ctx, c.Response().Writer)
}

func (s *Server) StatsHeaderFragment(c echo.Context) error {
	ctx := c.Request().Context()

	// Get latest indexed block
	latestBlockHeight, err := s.etl.GetDB().GetLatestIndexedBlock(ctx)
	if err != nil {
		s.logger.Warn("Failed to get latest block height", "error", err)
		latestBlockHeight = 0
	}

	// Get latest trusted block height from the trusted node
	trustedBlockHeight := int64(s.latestTrustedBlock.Load())

	// Calculate sync status - consider synced if within 60 blocks of the head
	const syncThreshold = 60
	var isSyncing bool
	var blockDelta int64
	syncProgressPercentage := float64(100) // Default to fully synced

	if trustedBlockHeight > 0 && latestBlockHeight > 0 {
		blockDelta = trustedBlockHeight - latestBlockHeight
		isSyncing = blockDelta > syncThreshold

		if isSyncing && trustedBlockHeight > 0 {
			syncProgressPercentage = (float64(latestBlockHeight) / float64(trustedBlockHeight)) * 100
			// Ensure percentage doesn't exceed 100
			if syncProgressPercentage > 100 {
				syncProgressPercentage = 100
			}
		} else {
			syncProgressPercentage = 100 // Fully synced
		}
	} else {
		// If we don't have trusted block info, assume synced
		isSyncing = false
		blockDelta = 0
		syncProgressPercentage = 100
	}

	// Get latest SLA rollup for BPS/TPS data
	var bps float64 = 0
	var avgBlockTime float32 = 0
	latestSlaRollup, err := s.etl.GetDB().GetLatestSlaRollup(ctx)
	if err != nil {
		s.logger.Debug("Failed to get latest SLA rollup", "error", err)
		// Fall back to default values
		bps = 0.5
		avgBlockTime = 2.0
	} else {
		bps = latestSlaRollup.Bps
		// Calculate average block time from BPS (if BPS > 0)
		if bps > 0 {
			avgBlockTime = float32(1.0 / bps)
		} else {
			avgBlockTime = 2.0 // Default 2 seconds
		}
	}

	// Get active validator count
	validatorCount, err := s.etl.GetDB().GetActiveValidatorCount(ctx)
	if err != nil {
		s.logger.Warn("Failed to get validator count", "error", err)
		validatorCount = 0
	}

	stats := &pages.DashboardStats{
		CurrentBlockHeight:  latestBlockHeight,
		ChainID:             s.etl.ChainID,
		BPS:                 bps,
		ValidatorCount:      validatorCount,
		AvgBlockTime:        avgBlockTime,
		IsSyncing:           isSyncing,
		LatestIndexedHeight: latestBlockHeight,
		LatestChainHeight:   trustedBlockHeight,
		BlockDelta:          blockDelta,
	}

	// Render the stats header fragment template
	fragment := pages.StatsHeaderFragment(stats, syncProgressPercentage)
	return fragment.Render(ctx, c.Response().Writer)
}

func (s *Server) TPSFragment(c echo.Context) error {
	ctx := c.Request().Context()

	// Get latest SLA rollup for TPS data
	var tps float64 = 0
	latestSlaRollup, err := s.etl.GetDB().GetLatestSlaRollup(ctx)
	if err != nil {
		s.logger.Debug("Failed to get latest SLA rollup", "error", err)
		// Fall back to default value
		tps = 0.1
	} else {
		tps = latestSlaRollup.Tps
	}

	// Get dashboard transaction stats from materialized view
	txStats, err := s.etl.GetDB().GetDashboardTransactionStats(ctx)
	if err != nil {
		s.logger.Warn("Failed to get dashboard transaction stats", "error", err)
		txStats = db.MvDashboardTransactionStat{}
	}

	stats := &pages.DashboardStats{
		TPS:                  tps,
		TotalTransactions30d: txStats.Transactions30d,
	}

	// Render the TPS fragment template
	fragment := pages.TPSFragment(stats)
	return fragment.Render(ctx, c.Response().Writer)
}

func (s *Server) TotalTransactionsFragment(c echo.Context) error {
	ctx := c.Request().Context()

	// Get dashboard transaction stats from materialized view
	txStats, err := s.etl.GetDB().GetDashboardTransactionStats(ctx)
	if err != nil {
		s.logger.Warn("Failed to get dashboard transaction stats", "error", err)
		txStats = db.MvDashboardTransactionStat{}
	}

	stats := &pages.DashboardStats{
		TotalTransactions:            txStats.TotalTransactions,
		TotalTransactions24h:         txStats.Transactions24h,
		TotalTransactionsPrevious24h: txStats.TransactionsPrevious24h,
	}

	// Render the total transactions fragment template
	fragment := pages.TotalTransactionsFragment(stats)
	return fragment.Render(ctx, c.Response().Writer)
}
