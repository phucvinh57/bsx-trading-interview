package engine_bench_test

import (
	"math/rand/v2"
	"net/http"
	"sync"
	"testing"
	"trading-bsx/cmd/api/server"
	"trading-bsx/internal/trade"
	"trading-bsx/pkg/db/models"
	"trading-bsx/pkg/testutil"
)

func Benchmark_PlaceOrders(b *testing.B) {
	b.Setenv("ENV", "test")
	s := server.New()
	defer s.Close()

	const minPrice = 100.0
	const maxPrice = 200.0
	const numOfUsers = 100

	wg := sync.WaitGroup{}

	for i := 1; i <= numOfUsers; i++ {
		wg.Add(1)
		go func(userId uint64) {
			defer wg.Done()
			client := testutil.NewClient(s)
			client.SetUser(userId)

			for j := 0; j < b.N; j++ {
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
		}(uint64(i))
	}
	wg.Wait()
}
