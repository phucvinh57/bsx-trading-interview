package engine_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
	"trading-bsx/cmd/api/server"
	"trading-bsx/internal/trade"
	"trading-bsx/pkg/db/models"
	"trading-bsx/pkg/db/mongodb"
	"trading-bsx/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func Test_PlaceOrders(t *testing.T) {
	t.Setenv("ENV", "test")
	s := server.New()
	defer s.Close()

	client := testutil.NewClient(s)
	var prices = []float64{100.5, 110.2, 120.3, 130.4, 140.5, 150.6, 160.7, 170.8, 180.9, 190.0}

	for _, price := range prices {
		for j := 1; j <= 2; j++ {
			client.SetUser(uint64(j))
			res := client.Request(&testutil.RequestOption{
				Method: http.MethodPost,
				URL:    "/orders",
				Body: trade.CreateOrder{
					Type:  models.BUY,
					Price: price,
				},
			})
			assert.Equal(t, http.StatusOK, res.Code)
		}
	}

	client.SetUser(1)
	res := client.Request(&testutil.RequestOption{
		Method: http.MethodGet,
		URL:    "/orders",
	})
	var orders = make([]models.Order, 0)
	json.NewDecoder(res.Body).Decode(&orders)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Len(t, orders, len(prices))
}

func Test_MatchBuyOrder(t *testing.T) {
	t.Setenv("ENV", "test")
	s := server.New()
	defer s.Close()

	client := testutil.NewClient(s)
	client.SetUser(1)

	var prices = []float64{100.5, 110.2, 120.3, 130.4, 140.5, 150.6, 160.7, 170.8, 180.9, 190.0}
	for _, price := range prices {
		client.Request(&testutil.RequestOption{
			Method: http.MethodPost,
			URL:    "/orders",
			Body: trade.CreateOrder{
				Type:  models.BUY,
				Price: price,
			},
		})
	}

	var matchPrices = []float64{190.0, 180.9}
	client.SetUser(2)
	for _, matchPrice := range matchPrices {
		res := client.Request(&testutil.RequestOption{
			Method: http.MethodPost,
			URL:    "/orders",
			Body: trade.CreateOrder{
				Type:  models.SELL,
				Price: 100.0,
			},
		})
		assert.Equal(t, http.StatusOK, res.Code)
		matchedOrder := models.Order{}
		err := json.NewDecoder(res.Body).Decode(&matchedOrder)
		assert.NoError(t, err)

		assert.Equal(t, matchPrice, matchedOrder.Price)
	}

	// Check order unmatched
	res := client.Request(&testutil.RequestOption{
		Method: http.MethodPost,
		URL:    "/orders",
		Body: trade.CreateOrder{
			Type:  models.SELL,
			Price: 500.0,
		},
	})
	assert.Equal(t, http.StatusOK, res.Code)
	orderId := res.Body.String()
	assert.Len(t, orderId, 24)

	numOfOrders, err := mongodb.Order.CountDocuments(context.Background(), bson.D{})
	assert.NoError(t, err)
	assert.Equal(t, len(prices)-len(matchPrices)+1, int(numOfOrders))
}

func Test_MatchSellOrder(t *testing.T) {
	t.Setenv("ENV", "test")
	s := server.New()
	defer s.Close()
	client := testutil.NewClient(s)
	client.SetUser(1)

	var prices = []float64{100.5, 110.2, 120.3, 130.4, 140.5, 150.6, 160.7, 170.8, 180.9, 190.0}
	for _, price := range prices {
		client.Request(&testutil.RequestOption{
			Method: http.MethodPost,
			URL:    "/orders",
			Body: trade.CreateOrder{
				Type:  models.SELL,
				Price: price,
			},
		})
	}

	var matchPrices = []float64{100.5, 110.2}
	client.SetUser(2)
	for _, matchPrice := range matchPrices {
		res := client.Request(&testutil.RequestOption{
			Method: http.MethodPost,
			URL:    "/orders",
			Body: trade.CreateOrder{
				Type:  models.BUY,
				Price: 140.0,
			},
		})
		assert.Equal(t, http.StatusOK, res.Code)
		matchedOrder := models.Order{}
		err := json.NewDecoder(res.Body).Decode(&matchedOrder)
		assert.NoError(t, err)

		assert.Equal(t, matchPrice, matchedOrder.Price)
	}

	// Check order unmatched
	res := client.Request(&testutil.RequestOption{
		Method: http.MethodPost,
		URL:    "/orders",
		Body: trade.CreateOrder{
			Type:  models.BUY,
			Price: 50.0,
		},
	})
	assert.Equal(t, http.StatusOK, res.Code)
	orderId := res.Body.String()
	assert.Len(t, orderId, 24)

	// Check total number of orders after matching
	numOfOrders, err := mongodb.Order.CountDocuments(context.Background(), bson.D{})
	assert.NoError(t, err)
	assert.Equal(t, len(prices)-len(matchPrices)+1, int(numOfOrders))
}

func Test_ExpiredOrder_ShouldNotMatch(t *testing.T) {
	t.Setenv("ENV", "test")
	s := server.New()
	defer s.Close()
	client := testutil.NewClient(s)
	client.SetUser(1)
	var gtt uint64 = 10
	client.Request(&testutil.RequestOption{
		Method: http.MethodPost,
		URL:    "/orders",
		Body: trade.CreateOrder{
			Type:  models.SELL,
			Price: 100.0,
			GTT:   &gtt,
		},
	})

	client.SetUser(2)
	time.Sleep(time.Duration(gtt) * time.Millisecond)
	res := client.Request(&testutil.RequestOption{
		Method: http.MethodPost,
		URL:    "/orders",
		Body: trade.CreateOrder{
			Type:  models.BUY,
			Price: 101.0,
		},
	})

	// Check order unmatched
	assert.Equal(t, http.StatusOK, res.Code)
	orderId := res.Body.String()
	assert.Len(t, orderId, 24)

	// Check total number of orders. User 1's order is expired, so num of his orders should be 0
	client.SetUser(1)
	res = client.Request(&testutil.RequestOption{
		Method: http.MethodGet,
		URL:    "/orders",
	})
	assert.Equal(t, http.StatusOK, res.Code)
	payload := res.Body.String()
	payload = strings.Trim(payload, "\n")
	assert.Equal(t, "[]", payload)
}
