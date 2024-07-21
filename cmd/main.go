package main

import (
	"trading-bsx/internal/middleware"
	"trading-bsx/internal/trade"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.Use(middleware.VerifyUser)

	order := e.Group("/orders")
	order.POST("", trade.PlaceOrder)

	e.Logger.Fatal(e.Start(":8080"))
}