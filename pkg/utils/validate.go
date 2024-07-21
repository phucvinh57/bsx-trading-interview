package utils

import "github.com/labstack/echo/v4"

func BindNValidate(c echo.Context, body interface{}) error {
	if err := c.Bind(body); err != nil {
		return err
	}

	if err := c.Validate(body); err != nil {
		return err
	}

	return nil
}
