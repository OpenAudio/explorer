package server

import (
	"github.com/AudiusProject/audiusd/pkg/console/templates/pages"
	"github.com/labstack/echo/v4"
)

func (s *Server) Content(c echo.Context) error {
	p := pages.Content()
	ctx := c.Request().Context()
	return p.Render(ctx, c.Response().Writer)
}

func (s *Server) stubRoute(c echo.Context) error {
	return c.String(200, "Hello, World!")
}
