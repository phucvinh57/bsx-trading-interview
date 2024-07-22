package trade

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"time"
	"trading-bsx/pkg/repository"
	"trading-bsx/pkg/utils"

	"github.com/labstack/echo/v4"
	"github.com/linxGnu/grocksdb"
)

type OrderType string

const (
	BUY  OrderType = "BUY"
	SELL OrderType = "SELL"
)

type CreateOrder struct {
	Type  OrderType `json:"type" validate:"required,oneof=BUY SELL"`
	Price float64   `json:"price" validate:"required,gt=0"`
	GTT   *uint64   `json:"gtt,omitempty" validate:"omitempty,gt=0"`
}

type Order struct {
	Key   []byte
	Value []byte
}

func NewOrder(orderType OrderType, price float64, userId uint64, gtt *uint64) Order {
	newOrder := Order{
		Key: newOrderKey(orderType, price, userId),
	}
	if gtt != nil {
		newOrder.Value = make([]byte, 8)
		binary.BigEndian.PutUint64(newOrder.Value, *gtt)
	}
	return newOrder
}

func (order Order) ToJSON() string {
	return ""
}

func newOrderKey(orderType OrderType, price float64, userId uint64) []byte {
	// 1 byte for order type, 16 bytes for price, 8 bytes for timestamp, 8 bytes for user ID
	key := make([]byte, 0, 33)
	if orderType == SELL {
		key = append(key, 1)
	} else {
		key = append(key, 0)
	}

	rawPrice := big.NewFloat(price)
	priceInt := big.NewInt(0)
	rawPrice.Mul(rawPrice, big.NewFloat(WEI18)).Int(priceInt)
	key = append(key, priceInt.Bytes()...)

	timestamp := time.Now().UnixNano()
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp))
	key = append(key, timestampBytes...)

	userIdBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(userIdBytes, userId)
	key = append(key, userIdBytes...)

	return key
}

const WEI18 = 1e18

func PlaceOrder(c echo.Context) error {
	body := CreateOrder{}
	if err := utils.BindNValidate(c, &body); err != nil {
		fmt.Println(err)
		return err
	}


	order := NewOrder(body.Type, body.Price, c.Get("userId").(uint64), body.GTT)

	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()
	err := repository.RocksDB.Put(wo, order.Key, order.Value)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, hex.EncodeToString(order.Key))
}
