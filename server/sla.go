package server

import (
	"strconv"

	"github.com/OpenAudio/explorer/templates/pages"
	"github.com/OpenAudio/explorer/db"
	"github.com/labstack/echo/v4"
)

func (s *Server) Rollups(c echo.Context) error {
	// Parse query parameters for pagination
	pageParam := c.QueryParam("page")
	countParam := c.QueryParam("count")

	page := int32(1) // default to page 1
	if pageParam != "" {
		if parsedPage, err := strconv.ParseInt(pageParam, 10, 32); err == nil && parsedPage > 0 {
			page = int32(parsedPage)
		}
	}

	count := int32(20) // default to 20 per page
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
