package trade

import (
	"encoding/base32"
	"fmt"
	"net/http"
	"trading-bsx/pkg/repository/rocksdb"
	"trading-bsx/pkg/utils"

	"github.com/labstack/echo/v4"
	"github.com/linxGnu/grocksdb"
)

type DeleteOrder struct {
	OrderKey string    `param:"order_key" validate:"required"`
	Type     OrderType `query:"type" validate:"required,oneof=BUY SELL"`
}

func CancelOrder(c echo.Context) error {
	req := DeleteOrder{}
	if err := utils.BindNValidate(c, &req); err != nil {
		fmt.Println(err)
		return err
	}

	userId := c.Get("userId").(uint64)
	var book *grocksdb.DB
	if req.Type == BUY {
		book = rocksdb.BuyOrder
	} else {
		book = rocksdb.SellOrder
	}
	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()

	orderKey, err := base32.StdEncoding.DecodeString(req.OrderKey)
	if err != nil {
		return err
	}	
	order := Order{ Type: req.Type }
	order.ParseKV(orderKey, nil)
	if order.UserId != userId {
		return echo.ErrNotFound
	}

	if err := book.Delete(wo, orderKey); err != nil {
		return err
	}

	return c.String(http.StatusOK, req.OrderKey)
}
