package engine_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"trading-bsx/cmd/api/server"
	"trading-bsx/internal/trade"
	"trading-bsx/pkg/db/models"
	"trading-bsx/pkg/testutil"

	"github.com/stretchr/testify/assert"
)

func Test_CancelOrder(t *testing.T) {
	t.Setenv("ENV", "test")
	s := server.New()
	defer s.Close()
	client := testutil.NewClient(s)
	client.SetUser(1)

	res := client.Request(&testutil.RequestOption{
		Method: http.MethodPost,
		URL:    "/orders",
		Body: trade.CreateOrder{
			Type:  models.SELL,
			Price: 100.0,
		},
	})
	assert.Equal(t, http.StatusOK, res.Code)
	orderId := res.Body.String()

	res = client.Request(&testutil.RequestOption{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/orders/%s", orderId),
	})
	assert.Equal(t, http.StatusOK, res.Code)

	// Check total number of orders. User 1's order is expired, so num of his orders should be 0
	res = client.Request(&testutil.RequestOption{
		Method: http.MethodGet,
		URL:    "/orders",
	})
	assert.Equal(t, http.StatusOK, res.Code)
	payload := res.Body.String()
	payload = strings.Trim(payload, "\n")
	assert.Equal(t, "[]", payload)
}

func Test_CancelOrder_MustNotMatch_NewOrder(t *testing.T) {
	t.Setenv("ENV", "test")
	s := server.New()
	defer s.Close()
	client := testutil.NewClient(s)
	client.SetUser(1)

	res := client.Request(&testutil.RequestOption{
		Method: http.MethodPost,
		URL:    "/orders",
		Body: trade.CreateOrder{
			Type:  models.SELL,
			Price: 100.0,
		},
	})
	assert.Equal(t, http.StatusOK, res.Code)
	orderId := res.Body.String()

	res = client.Request(&testutil.RequestOption{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/orders/%s", orderId),
	})
	assert.Equal(t, http.StatusOK, res.Code)

	client.SetUser(2)
	res = client.Request(&testutil.RequestOption{
		Method: http.MethodPost,
		URL:    "/orders",
		Body: trade.CreateOrder{
			Type:  models.BUY,
			Price: 101.0,
		},
	})
	assert.Equal(t, http.StatusOK, res.Code)
	orderId = res.Body.String()
	assert.Len(t, orderId, 24) // New order's id returned, not match result
}