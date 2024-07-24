package engine_bench_test

import (
	"math/rand/v2"
	"net/http"
	"testing"
	"trading-bsx/cmd/api/server"
	"trading-bsx/internal/trade"
	"trading-bsx/pkg/db/models"
	"trading-bsx/pkg/testutil"
)

func Benchmark_PlaceOnlyOneOrderType(b *testing.B) {
	b.Setenv("ENV", "test")
	s := server.New()
	defer s.Close()

	const minPrice = 100.0
	const maxPrice = 200.0

	client := testutil.NewClient(s)

	for j := 0; j < b.N; j++ {
		client.SetUser(rand.Uint64N(200))
		price := minPrice + rand.Float64()*(maxPrice-minPrice)
		client.Request(&testutil.RequestOption{
			Method: http.MethodPost,
			URL:    "/orders",
			Body: trade.CreateOrder{
				Type:  models.BUY,
				Price: price,
			},
		})
	}
}

func Benchmark_PlaceRandomBuyNSellOrders(b *testing.B) {
	b.Setenv("ENV", "test")
	s := server.New()
	defer s.Close()

	const minPrice = 100.0
	const maxPrice = 200.0

	client := testutil.NewClient(s)

	for j := 0; j < b.N; j++ {
		var orderType models.OrderType
		client.SetUser(rand.Uint64N(200))
		price := minPrice + rand.Float64()*(maxPrice-minPrice)

		if rand.IntN(2) == 0 {
			orderType = models.BUY
		} else {
			orderType = models.SELL
		}
		client.Request(&testutil.RequestOption{
			Method: http.MethodPost,
			URL:    "/orders",
			Body: trade.CreateOrder{
				Type:  orderType,
				Price: price,
			},
		})
	}
}
