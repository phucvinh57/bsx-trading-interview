package trade

import (
	"net/http"
	"trading-bsx/pkg/repository/rocksdb"

	"github.com/labstack/echo/v4"
	"github.com/linxGnu/grocksdb"
)

func GetOrders(c echo.Context) error {
	ro := grocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	sellIt := rocksdb.SellOrder.NewIterator(ro)
	defer sellIt.Close()

	orders := make([]Order, 0)
	for sellIt.SeekToFirst(); sellIt.Valid(); sellIt.Next() {
		order := Order{ Type: SELL }
		order.ParseKV(sellIt.Key().Data(), sellIt.Value().Data())
		orders = append(orders, order)
	}

	buyIt := rocksdb.BuyOrder.NewIterator(ro)
	defer buyIt.Close()
	for buyIt.SeekToLast(); buyIt.Valid(); buyIt.Prev() {
		order := Order{ Type: BUY }
		order.ParseKV(buyIt.Key().Data(), buyIt.Value().Data())
		orders = append(orders, order)
	}

	return c.JSON(http.StatusOK, orders)
}
