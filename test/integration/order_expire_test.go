package engine_test

import (
	"net/http"
	"strings"
	"testing"
	"time"
	"trading-bsx/cmd/api/server"
	"trading-bsx/internal/trade"
	"trading-bsx/pkg/db/models"
	"trading-bsx/pkg/testutil"

	"github.com/stretchr/testify/assert"
)

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
