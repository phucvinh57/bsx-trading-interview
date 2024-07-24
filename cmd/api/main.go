package main

import (
	"trading-bsx/cmd/api/server"

	"github.com/rs/zerolog/log"
)

func main() {
	s := server.New()
	err := s.Start(":8080")
	if err != nil {
		log.Err(err).Send()
	}
}
