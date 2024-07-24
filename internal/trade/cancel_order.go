package trade

import (
	"encoding/base32"
	"fmt"
	"net/http"
	"trading-bsx/pkg/db/models"
	"trading-bsx/pkg/db/mongodb"
	"trading-bsx/pkg/db/rocksdb"
	"trading-bsx/pkg/utils"

	"github.com/labstack/echo/v4"
	"github.com/linxGnu/grocksdb"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
)

type DeleteOrder struct {
	OrderKey string    `param:"order_key" validate:"required"`
	Type     models.OrderType `query:"type" validate:"required,oneof=BUY SELL"`
}

func CancelOrder(c echo.Context) error {
	req := DeleteOrder{}
	if err := utils.BindNValidate(c, &req); err != nil {
		fmt.Println(err)
		return err
	}

	userId := c.Get("userId").(uint64)
	var book *grocksdb.DB
	if req.Type == models.BUY {
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
	order := models.Order{ Type: req.Type }
	order.ParseKV(orderKey, nil)
	if order.UserId != userId {
		return echo.ErrNotFound
	}

	if err := book.Delete(wo, orderKey); err != nil {
		return err
	}
	result, err := mongodb.Order.DeleteOne(c.Request().Context(), bson.M{
		"key": req.OrderKey,
	})
	if err != nil {
		return err
	}
	log.Info().Interface("result", result).Msg("Cancel order")

	return c.String(http.StatusOK, req.OrderKey)
}
