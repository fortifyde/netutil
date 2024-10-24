package main

import (
	"fmt"
	"os"

	"github.com/fortifyde/netutil/internal/functions/configuration"
	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/ui"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	primitive := tview.NewBox()

	cfg, err := configuration.ReadWriteConfig(app, primitive)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read/write config: %v\n", err)
		os.Exit(1)
	}

	err = logger.Init(cfg.WorkingDirectory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("Main program started")

	if err := ui.RunApp(); err != nil {
		logger.Error("Application error: %v", err)
		os.Exit(1)
	}
}
