package main

import (
	"os"
	"trading-bsx/internal/middleware"
	"trading-bsx/internal/trade"
	"trading-bsx/pkg/db/mongodb"
	"trading-bsx/pkg/db/rocksdb"
	"trading-bsx/pkg/utils"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stdout,
	})
	rocksdb.Init()
	mongodb.Init("mongodb://localhost:27017")

	e := echo.New()
	e.Validator = utils.NewValidator()
	e.Use(middleware.VerifyUser)

	order := e.Group("/orders")
	order.GET("", trade.GetOrders)
	order.POST("", trade.PlaceOrder)
	order.DELETE("/:order_key", trade.CancelOrder)

	log.Err(e.Start(":8080")).Send()
}
