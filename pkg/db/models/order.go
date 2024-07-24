package models

import (
	"encoding/base32"
	"encoding/binary"
	"math/big"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderType string

const (
	BUY  OrderType = "BUY"
	SELL OrderType = "SELL"
)

type Order struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserId    uint64             `json:"userId" bson:"user_id"`
	Type      OrderType          `json:"type" bson:"type"`
	Price     float64            `json:"price" bson:"price"`
	ExpiredAt *uint64            `json:"expiredAt,omitempty" bson:"expired_at,omitempty"`
	Timestamp uint64             `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	Key       string             `json:"key,omitempty" bson:"key,omitempty"`
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
	order.Timestamp = ts
	order.UserId = binary.BigEndian.Uint64(key[24:32])

	if len(value) > 0 {
		exp := binary.BigEndian.Uint64(value)
		order.ExpiredAt = &exp
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

	if order.Type == BUY {
		order.Timestamp = ^order.Timestamp
	}
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, order.Timestamp)

	copy(key[16:24], timestampBytes)

	userIdBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(userIdBytes, order.UserId)
	copy(key[24:32], userIdBytes)

	value := make([]byte, 8)
	if order.ExpiredAt != nil {
		binary.BigEndian.PutUint64(value, *order.ExpiredAt)
	}
	return key, value
}

const WEI18 = 1e18
