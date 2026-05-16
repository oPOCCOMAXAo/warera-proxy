package main

import (
	"log"

	"github.com/opoccomaxao/warera-proxy/pkg/app"
)

func main() {
	err := app.Run()
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
