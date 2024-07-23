package trade

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"time"
	"trading-bsx/pkg/repository/rocksdb"
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
	price := new(big.Int).SetBytes(key[:16])
	priceFloat := big.NewFloat(0).SetInt(price)
	priceFloat.Quo(priceFloat, big.NewFloat(WEI18))
	order.Price, _ = priceFloat.Float64()

	order.Timestamp = int64(binary.BigEndian.Uint64(key[16:24]))
	order.UserId = binary.BigEndian.Uint64(key[24:32])
}

func (order *Order) ToKVBytes() ([]byte, []byte) {
	// 16 bytes for price, 8 bytes for timestamp, 8 bytes for user ID
	key := make([]byte, 32)

	rawPrice := big.NewFloat(order.Price)
	priceInt := big.NewInt(0)
	rawPrice.Mul(rawPrice, big.NewFloat(WEI18)).Int(priceInt)
	copy(key[16-len(priceInt.Bytes()):], priceInt.Bytes())

	timestamp := time.Now().UnixNano()
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp))
	copy(key[16:24], timestampBytes)

	userIdBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(userIdBytes, order.UserId)
	copy(key[24:32], userIdBytes)

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

	var opponentBook *grocksdb.DB
	var orderBook *grocksdb.DB
	var opponentType OrderType
	if order.Type == BUY {
		opponentBook = rocksdb.SellOrder
		orderBook = rocksdb.BuyOrder
		opponentType = SELL
	} else {
		opponentBook = rocksdb.BuyOrder
		orderBook = rocksdb.SellOrder
		opponentType = BUY
	}

	it := opponentBook.NewIterator(ro)
	defer it.Close()

	it.Seek(orderKey)
	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()

	for it.Valid() {
		k, v := it.Key().Data(), it.Value().Data()
		matchOrder := Order{
			Type: opponentType,
		}
		matchOrder.ParseKV(k, v)
		isSameUser := order.UserId == matchOrder.UserId

		if isSameUser {
			it.Next()
			continue
		}

		gtt := matchOrder.GTT
		if gtt == nil || *gtt == 0 {
			break
		}

		isExpired := time.Now().UnixNano() > matchOrder.Timestamp+int64(*gtt)
		if isExpired {
			if err := opponentBook.Delete(wo, k); err != nil {
				return err
			}
			it.Next()
			continue
		}

		break
	}

	if !it.Valid() { // No matching order found
		err := orderBook.Put(wo, orderKey, gtt)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, hex.EncodeToString(orderKey))
	}

	// Buy -> Get smallest sell order
	// Sell -> Get biggest buy order

	// Matching order found
	matchOrder := Order{}
	matchOrder.ParseKV(it.Key().Data(), it.Value().Data())

	// err := opponentBook.Delete(wo, it.Key().Data())
	// if err != nil {
	// 	return err
	// }

	return c.JSON(http.StatusOK, matchOrder)
}
