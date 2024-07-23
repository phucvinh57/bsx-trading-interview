package trade

import (
	"encoding/binary"
	"encoding/base32"
	"fmt"
	"math/big"
	"net/http"
	"time"
	"trading-bsx/pkg/repository/rocksdb"
	"trading-bsx/pkg/utils"

	"github.com/labstack/echo/v4"
	"github.com/linxGnu/grocksdb"
	"github.com/rs/zerolog/log"
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
	Timestamp int64     `json:"timestamp,omitempty"`
	Key       string    `json:"key,omitempty"`
}

func (order *Order) ParseKV(key []byte, value []byte) {
	price := new(big.Int).SetBytes(key[:16])
	priceFloat := big.NewFloat(0).SetInt(price)
	priceFloat.Quo(priceFloat, big.NewFloat(WEI18))
	order.Price, _ = priceFloat.Float64()

	ts := binary.BigEndian.Uint64(key[16:24])
	if order.Type == BUY {
		ts = ^ts
	}
	order.Timestamp = int64(ts)
	order.UserId = binary.BigEndian.Uint64(key[24:32])

	if len(value) > 0 {
		gtt := binary.BigEndian.Uint64(value)
		order.GTT = &gtt
	}

	order.Key = base32.StdEncoding.EncodeToString(key)
}

func (order *Order) ToKVBytes() ([]byte, []byte) {
	// 16 bytes for price, 8 bytes for timestamp, 8 bytes for user ID
	key := make([]byte, 32)

	rawPrice := big.NewFloat(order.Price)
	priceInt := big.NewInt(0)
	rawPrice.Mul(rawPrice, big.NewFloat(WEI18)).Int(priceInt)
	copy(key[16-len(priceInt.Bytes()):], priceInt.Bytes())

	timestamp := uint64(time.Now().UnixNano())
	if order.Type == BUY {
		timestamp = ^timestamp
	}
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, timestamp)

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

func getMatchBuyOrder(order *Order) ([]byte, *Order) {
	// Sell -> Get biggest buy order -> Seek from the last item in the list
	ro := grocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	it := rocksdb.BuyOrder.NewIterator(ro)
	defer it.Close()

	it.SeekToLast()
	for it.Valid() {
		k, v := it.Key().Data(), it.Value().Data()
		matchOrder := Order{
			Type: BUY,
		}
		matchOrder.ParseKV(k, v)
		if matchOrder.UserId == order.UserId {
			it.Prev()
			continue
		}
		gtt := matchOrder.GTT
		if gtt != nil && *gtt != 0 {
			if time.Now().UnixNano() > matchOrder.Timestamp+int64(*gtt)*int64(time.Second) {
				wo := grocksdb.NewDefaultWriteOptions()
				defer wo.Destroy()
				if err := rocksdb.BuyOrder.Delete(wo, k); err != nil {
					fmt.Println(err)
				}
				it.Prev()
				continue
			}
		}
		if matchOrder.Price >= order.Price {
			return k, &matchOrder
		}

		// The biggest buy order is smaller than the current order, so no need to continue
		return nil, nil
	}
	return nil, nil
}

func getMatchSellOrder(order *Order) ([]byte, *Order) {
	// Buy -> Get smallest sell order -> Seek from the first item in the list
	ro := grocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	it := rocksdb.SellOrder.NewIterator(ro)
	defer it.Close()

	it.SeekToFirst()
	for it.Valid() {
		k, v := it.Key().Data(), it.Value().Data()
		matchOrder := Order{
			Type: SELL,
		}
		matchOrder.ParseKV(k, v)
		if matchOrder.UserId == order.UserId {
			it.Next()
			continue
		}
		gtt := matchOrder.GTT
		if gtt != nil && *gtt != 0 {
			if time.Now().UnixNano() > matchOrder.Timestamp+int64(*gtt)*int64(time.Second) {
				wo := grocksdb.NewDefaultWriteOptions()
				defer wo.Destroy()
				if err := rocksdb.SellOrder.Delete(wo, k); err != nil {
					fmt.Println(err)
				}
				it.Next()
				continue
			}
		}
		if matchOrder.Price <= order.Price {
			return k, &matchOrder
		}

		// The smallest sell order is bigger than the current order, so no need to continue
		return nil, nil
	}
	return nil, nil
}

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

	var book *grocksdb.DB
	var opponentBook *grocksdb.DB
	var matchOrder *Order
	var matchOrderKey []byte

	if order.Type == BUY {
		book = rocksdb.BuyOrder
		opponentBook = rocksdb.SellOrder
		matchOrderKey, matchOrder = getMatchSellOrder(&order)
	} else {
		book = rocksdb.SellOrder
		opponentBook = rocksdb.BuyOrder
		matchOrderKey, matchOrder = getMatchBuyOrder(&order)
	}

	log.Info().Interface("order", order).Msg("Place order")
	log.Info().Interface("matchOrder", matchOrder).Msg("Match order")

	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()

	if matchOrder != nil {
		if err := opponentBook.Delete(wo, matchOrderKey); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, matchOrder)
	}

	orderKey, orderValue := order.ToKVBytes()
	if err := book.Put(wo, orderKey, orderValue); err != nil {
		return err
	}

	order.Key = base32.StdEncoding.EncodeToString(orderKey)
	return c.String(http.StatusOK, order.Key)
}
