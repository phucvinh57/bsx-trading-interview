package trade

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetOrders(c echo.Context) error {
	return c.String(http.StatusOK, "GetOrders")
}
