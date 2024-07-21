package trade

import (
	"fmt"
	"net/http"
	"trading-bsx/pkg/utils"

	"github.com/labstack/echo/v4"
)

type OrderType string

const (
	BUY  OrderType = "BUY"
	SELL OrderType = "SELL"
)

type Order struct {
	Type  string  `json:"type" validate:"required,oneof=BUY SELL"`
	Price float64 `json:"price" validate:"required,gt=0"`
	GTT   *uint64 `json:"gtt,omitempty" validate:"omitempty,gt=0"`
}

func PlaceOrder(c echo.Context) error {
	order := Order{}
	if err := utils.BindNValidate(c, &order); err != nil {
		fmt.Println(err)
		return err
	}
	return c.JSON(http.StatusOK, order)
}
