package trade

import (
	"net/http"
	"time"
	"trading-bsx/pkg/db/models"
	"trading-bsx/pkg/db/mongodb"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
)

func GetOrders(c echo.Context) error {
	userId := c.Get("userId").(uint64)
	reqCtx := c.Request().Context()
	orders := make([]models.Order, 0)
	ts := uint64(time.Now().UnixNano())
	filter := bson.M{
		"$and": []bson.M{
			{"user_id": userId},
			{
				"$or": []bson.M{
					{"expired_at": bson.M{"$gte": ts}},
					{"expired_at": 0},
					{"expired_at": bson.M{"$exists": false}},
				},
			},
		},
	}
	cursor, err := mongodb.Order.Find(reqCtx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(reqCtx)
	err = cursor.All(reqCtx, &orders)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, orders)
}
