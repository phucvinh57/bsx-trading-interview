package trade

import "github.com/labstack/echo/v4"

func PlaceOrder(c echo.Context) error {
	return c.String(200, "Order placed")
}