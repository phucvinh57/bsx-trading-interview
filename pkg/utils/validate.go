package utils

import (
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewValidator(v *validator.Validate) *CustomValidator {
	return &CustomValidator{validator: v}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func BindNValidate(c echo.Context, body interface{}) error {
	if err := c.Bind(body); err != nil {
		return err
	}
	if err := c.Validate(body); err != nil {
		return err
	}
	return nil
}
