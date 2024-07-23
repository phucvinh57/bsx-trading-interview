package trade

import (
	"net/http"
	"trading-bsx/pkg/repository"

	"github.com/labstack/echo/v4"
	"github.com/linxGnu/grocksdb"
)

func GetOrders(c echo.Context) error {
	ro := grocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	it := repository.RocksDB.NewIterator(ro)

	orders := make([]Order, 0)
	for it.SeekToFirst(); it.Valid(); it.Next() {
		order := Order{}
		order.ParseKV(it.Key().Data(), it.Value().Data())
		orders = append(orders, order)
	}
	return c.JSON(http.StatusOK, orders)
}
