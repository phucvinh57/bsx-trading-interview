package trade

import (
	"fmt"
	"net/http"
	"trading-bsx/pkg/utils"

	"github.com/labstack/echo/v4"
)

type DeleteOrder struct {
	OrderId string `param:"order_id" validate:"required"`
}

func CancelOrder(c echo.Context) error {
	req := DeleteOrder{}
	if err := utils.BindNValidate(c, &req); err != nil {
		fmt.Println(err)
		return err
	}
	return c.String(http.StatusOK, req.OrderId)
}
