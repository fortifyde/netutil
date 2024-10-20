package main

import (
	"log"

	"github.com/fortifyde/netutil/internal/functions"
	"github.com/fortifyde/netutil/internal/ui"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	primitive := tview.NewBox() // or another tview.Primitive as needed

	functions.ReadWriteConfig(app, primitive)
	if err := ui.RunApp(); err != nil {
		log.Fatal(err)
	}
}
