package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AudiusProject/audiusd/pkg/etl"
	"github.com/AudiusProject/audiusd/pkg/etl/db"
	"github.com/labstack/echo/v4"
)

type SSEEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

const sseConnectionTTL = 1 * time.Minute

func (s *Server) LiveEventsSSE(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)

	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		return nil
	}

	flusher.Flush()

	// Subscribe to both block and play events from ETL pubsub
	blockCh := s.etl.GetBlockPubsub().Subscribe(etl.BlockTopic, 10)
	playCh := s.etl.GetPlayPubsub().Subscribe(etl.PlayTopic, 10)

	// Ensure cleanup on connection close
	defer func() {
		s.etl.GetBlockPubsub().Unsubscribe(etl.BlockTopic, blockCh)
		s.etl.GetPlayPubsub().Unsubscribe(etl.PlayTopic, playCh)
	}()

	// Throttle state for block events
	var (
		latestBlock    *db.EtlBlock
		lastSentHeight int64
		blockTicker    = time.NewTicker(1 * time.Second)
	)
	defer blockTicker.Stop()

	flusher.Flush()

	timeout := time.After(sseConnectionTTL)

	for {
		select {
		case <-c.Request().Context().Done():
			return nil

		case <-timeout:
			return nil

		case blockEvent := <-blockCh:
			if blockEvent != nil {
				latestBlock = blockEvent
			}

		case <-blockTicker.C:
			if latestBlock != nil && latestBlock.BlockHeight > lastSentHeight {
				// Send block event
				blockEvent := SSEEvent{
					Event: "block",
					Data: map[string]interface{}{
						"height":   latestBlock.BlockHeight,
						"proposer": latestBlock.ProposerAddress,
						"time":     latestBlock.BlockTime.Time.Format(time.RFC3339),
					},
				}
				eventData, _ := json.Marshal(blockEvent)
				fmt.Fprintf(c.Response(), "data: %s\n\n", string(eventData))
				lastSentHeight = latestBlock.BlockHeight
				flusher.Flush()
			}

		case play := <-playCh:
			if play != nil {
				// Get coordinates for the play location
				if play.City != "" && play.Region != "" && play.Country != "" {
					if latLong, err := s.etl.GetLocationDB().GetLatLong(c.Request().Context(), play.City, play.Region, play.Country); err == nil {
						lat := latLong.Latitude
						lng := latLong.Longitude
						// Send play event with coordinates
						playEvent := SSEEvent{
							Event: "play",
							Data: map[string]interface{}{
								"lat":       lat,
								"lng":       lng,
								"timestamp": time.Now().Format(time.RFC3339),
								"duration":  5, // Default 5 seconds for animation
							},
						}
						eventData, _ := json.Marshal(playEvent)
						fmt.Fprintf(c.Response(), "data: %s\n\n", string(eventData))
						flusher.Flush()
					}
				}
			}

		}
	}
}
