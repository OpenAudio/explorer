package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"connectrpc.com/connect"
	"github.com/OpenAudio/explorer/assets"
	"github.com/OpenAudio/explorer/config"
	"github.com/OpenAudio/explorer/db"
	corev1 "github.com/OpenAudio/go-openaudio/pkg/api/core/v1"
	"github.com/OpenAudio/go-openaudio/pkg/sdk"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type Server struct {
	config      *config.Config
	e           *echo.Echo
	db          *db.Queries
	logger      *slog.Logger
	trustedNode *sdk.OpenAudioSDK
	chainID     string

	latestTrustedBlock atomic.Int64
	lastRefreshTime    atomic.Int64  // Unix timestamp of last refresh
	refreshInterval    time.Duration // How often to refresh
}

func New(config *config.Config) *Server {
	e := echo.New()
	e.HideBanner = true

	return &Server{
		e:               e,
		logger:          slog.Default(),
		config:          config,
		refreshInterval: 10 * time.Second,
	}
}

func (s *Server) Initialize() error {
	// Initialize database connection
	connString := fmt.Sprintf("postgres://postgres:postgres@localhost:5444/audiusd?sslmode=disable")
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("failed to create database connection pool: %w", err)
	}
	s.db = db.New(pool)

	// start refresher
	go s.refreshTrustedBlock()

	e := s.e
	e.HideBanner = true

	// Add environment context middleware
	envMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Add environment to the request context
			ctx := context.WithValue(c.Request().Context(), "env", s.config.Environment)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}

	// Add cache control middleware for static assets
	cacheControl := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			// Only apply caching to image files
			if strings.HasPrefix(path, "/assets/") && (strings.HasSuffix(path, ".svg") || strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") || strings.HasSuffix(path, ".gif")) {
				c.Response().Header().Set("Cache-Control", "public, max-age=604800") // Cache for 1 week
			}
			return next(c)
		}
	}

	cssHandler := echo.MustSubFS(assets.CSS, "css")
	imagesHandler := echo.MustSubFS(assets.Images, "images")
	jsHandler := echo.MustSubFS(assets.JS, "js")
	e.StaticFS("/assets/css", cssHandler)
	e.StaticFS("/assets/images", imagesHandler)
	e.StaticFS("/assets/js", jsHandler)

	// Apply middlewares
	e.Use(cacheControl)
	e.Use(envMiddleware)

	// Routes
	e.GET("/", s.Dashboard)

	e.GET("/validators", s.Validators)
	e.GET("/validator/:address", s.Validator)
	e.GET("/validators/uptime", s.ValidatorsUptime)
	e.GET("/validators/uptime/:rollupid", s.ValidatorsUptimeByRollup)

	e.GET("/rollups", s.Rollups)

	e.GET("/blocks", s.Blocks)
	e.GET("/block/:height", s.Block)

	e.GET("/transactions", s.Transactions)
	e.GET("/transaction/:hash", s.Transaction)

	e.GET("/account/:address", s.Account)
	e.GET("/account/:address/transactions", s.stubRoute)
	e.GET("/account/:address/uploads", s.stubRoute)
	e.GET("/account/:address/releases", s.stubRoute)

	e.GET("/content", s.Content)
	e.GET("/content/:address", s.Content)

	e.GET("/release/:address", s.stubRoute)

	e.GET("/search", s.Search)

	// SSE endpoints
	e.GET("/sse/events", s.LiveEventsSSE)

	// HTMX Fragment routes
	e.GET("/fragments/stats-header", s.StatsHeaderFragment)
	e.GET("/fragments/tps", s.TPSFragment)
	e.GET("/fragments/total-transactions", s.TotalTransactionsFragment)

	return nil
}

func (s *Server) Start() error {
	return s.e.Start(":3000")
}

func (s *Server) Stop() {
	s.e.Shutdown(context.Background())
}

// refreshTrustedBlock refreshes the trusted block height every 10 seconds
func (s *Server) refreshTrustedBlock() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		info, err := s.trustedNode.Core.GetNodeInfo(context.Background(), &connect.Request[corev1.GetNodeInfoRequest]{})
		if err != nil {
			s.logger.Warn("Failed to refresh node info", "error", err)
			// Use the cached value if refresh fails
		} else {
			s.latestTrustedBlock.Store(info.Msg.CurrentHeight)
			s.lastRefreshTime.Store(time.Now().Unix())
		}
	}
}
