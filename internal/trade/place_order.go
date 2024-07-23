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
	UserId    uint64    `json:"userId"`
	Type      OrderType `json:"type"`
	Price     float64   `json:"price"`
	GTT       *uint64   `json:"gtt,omitempty"`
	Timestamp int64     `json:"timestamp"`
}

func (order *Order) ParseKV(key []byte, value []byte) {
	// 1 byte for order type, 16 bytes for price, 8 bytes for timestamp, 8 bytes for user ID
	order.Type = BUY
	if key[0] == 1 {
		order.Type = SELL
	}
	price := new(big.Int).SetBytes(key[1:17])
	priceFloat := big.NewFloat(0).SetInt(price)
	priceFloat.Quo(priceFloat, big.NewFloat(WEI18))
	order.Price, _ = priceFloat.Float64()

	order.Timestamp = int64(binary.BigEndian.Uint64(key[17:25]))
	order.UserId = binary.BigEndian.Uint64(key[25:33])
}

func (order *Order) ToKVBytes() ([]byte, []byte) {
	// 1 byte for order type, 16 bytes for price, 8 bytes for timestamp, 8 bytes for user ID
	key := make([]byte, 33)
	if order.Type == SELL {
		key[0] = 1
	} else {
		key[0] = 0
	}

	rawPrice := big.NewFloat(order.Price)
	priceInt := big.NewInt(0)
	rawPrice.Mul(rawPrice, big.NewFloat(WEI18)).Int(priceInt)
	copy(key[17-len(priceInt.Bytes()):], priceInt.Bytes())
	key = append(key, priceInt.Bytes()...)

	timestamp := time.Now().UnixNano()
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp))
	copy(key[17:25], timestampBytes)

	userIdBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(userIdBytes, order.UserId)
	copy(key[25:33], userIdBytes)

	value := make([]byte, 8)
	if order.GTT != nil {
		binary.BigEndian.PutUint64(value, *order.GTT)
	}
	return key, value
}

const WEI18 = 1e18

func PlaceOrder(c echo.Context) error {
	body := CreateOrder{}
	if err := utils.BindNValidate(c, &body); err != nil {
		fmt.Println(err)
		return err
	}

	order := Order{
		UserId: c.Get("userId").(uint64),
		Type:   body.Type,
		Price:  body.Price,
		GTT:    body.GTT,
	}

	orderKey, gtt := order.ToKVBytes()

	ro := grocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	ro.SetFillCache(false)

	it := repository.RocksDB.NewIterator(ro)
	defer it.Close()

	it.Seek(orderKey)

	if !it.Valid() { // No matching order found
		wo := grocksdb.NewDefaultWriteOptions()
		defer wo.Destroy()

		err := repository.RocksDB.Put(wo, orderKey, gtt)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, hex.EncodeToString(orderKey))
	}

	// Matching order found
	matchOrder := Order{}
	matchOrder.ParseKV(it.Key().Data(), it.Value().Data())

	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()
	err := repository.RocksDB.Delete(wo, it.Key().Data())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, matchOrder)
}
