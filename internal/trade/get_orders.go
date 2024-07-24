package trade

import (
	"net/http"
	"trading-bsx/pkg/db/models"
	"trading-bsx/pkg/db/rocksdb"

	"github.com/labstack/echo/v4"
	"github.com/linxGnu/grocksdb"
)

func GetOrders(c echo.Context) error {
	ro := grocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	sellIt := rocksdb.SellOrder.NewIterator(ro)
	defer sellIt.Close()

	orders := make([]models.Order, 0)
	for sellIt.SeekToFirst(); sellIt.Valid(); sellIt.Next() {
		order := models.Order{Type: models.SELL}
		order.ParseKV(sellIt.Key().Data(), sellIt.Value().Data())
		orders = append(orders, order)
	}

	buyIt := rocksdb.BuyOrder.NewIterator(ro)
	defer buyIt.Close()
	for buyIt.SeekToLast(); buyIt.Valid(); buyIt.Prev() {
		order := models.Order{Type: models.BUY}
		order.ParseKV(buyIt.Key().Data(), buyIt.Value().Data())
		orders = append(orders, order)
	}

	return c.JSON(http.StatusOK, orders)
}
