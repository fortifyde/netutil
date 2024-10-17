package main

import (
	"log"

	"github.com/fortifyde/netutil/internal/ui"
)

func main() {
	if err := ui.RunApp(); err != nil {
		log.Fatal(err)
	}
}
