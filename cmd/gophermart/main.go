package main

import (
	"github.com/rookgm/gophermart/config"
	"log"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
}
