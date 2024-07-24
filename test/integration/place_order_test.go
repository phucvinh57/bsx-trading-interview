package engine_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"trading-bsx/cmd/api/server"
	"trading-bsx/internal/trade"
	"trading-bsx/pkg/db/models"
	"trading-bsx/pkg/testutil"

	"github.com/stretchr/testify/assert"
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
