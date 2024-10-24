package main

import (
	"fmt"
	"os"

	"github.com/fortifyde/netutil/internal/functions/configuration"
	"github.com/fortifyde/netutil/internal/logger"
	"github.com/fortifyde/netutil/internal/ui"
	"github.com/fortifyde/netutil/internal/uiutil"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	pages := tview.NewPages()

	uiutil.InitializeModalManager(app, pages, nil)

	// read and write configuration
	cfg, err := configuration.ReadWriteConfig(app, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read/write config: %v\n", err)
		os.Exit(1)
	}

	// initialize and defer close the logger
	err = logger.Init(cfg.WorkingDirectory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("Main program started")

	// run the UI application
	if err := ui.RunApp(app, pages, nil); err != nil {
		logger.Error("Application error: %v", err)
		os.Exit(1)
	}
}
