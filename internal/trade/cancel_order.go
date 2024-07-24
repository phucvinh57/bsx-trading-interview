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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeleteOrder struct {
	OrderId primitive.ObjectID    `param:"order_id" validate:"required"`
}

func CancelOrder(c echo.Context) error {
	req := DeleteOrder{}
	if err := utils.BindNValidate(c, &req); err != nil {
		fmt.Println(err)
		return err
	}
	userId := c.Get("userId").(uint64)

	order := models.Order{}
	if err := mongodb.Order.FindOneAndDelete(c.Request().Context(), bson.M{
		"_id": req.OrderId,
		"user_id": userId,
	}).Decode(&order); err != nil {
		return err
	}

	var book *grocksdb.DB
	if order.Type == models.BUY {
		book = rocksdb.BuyOrder
	} else {
		book = rocksdb.SellOrder
	}
	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()

	orderKey, _ := base32.StdEncoding.DecodeString(order.Key)

	mutex.Lock()
	defer mutex.Unlock()
	if err := book.Delete(wo, orderKey); err != nil {
		return err
	}

	log.Info().Interface("order", order).Msg("Cancel order")
	return c.String(http.StatusOK, order.ID.Hex())
}
