package middleware

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

func VerifyUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the user ID from the JWT
		authHeader := c.Request().Header.Get("Authorization")
		if len(authHeader) == 0 {
			return echo.ErrUnauthorized
		}

		userId, err := strconv.ParseUint(authHeader, 10, 64)
		if err != nil {
			return echo.ErrUnauthorized
		}

		c.Set("userId", userId)

		return next(c)
	}
}
