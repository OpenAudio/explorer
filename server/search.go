package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) Search(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"results": []interface{}{},
		})
	}

	// TODO: Implement search using database queries
	return c.JSON(http.StatusOK, map[string]interface{}{
		"results": []interface{}{},
	})
}
