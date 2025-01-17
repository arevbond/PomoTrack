package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/arevbond/PomoTrack/config"
)

func main() {
	logger, err := initLogger("app.log")
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.Init()
	if err != nil {
		logger.Error("can't initialization config", slog.Any("error", err))
		os.Exit(1)
	}

	app := NewApplication(logger, cfg)

	// non blocking operation using goroutins
	app.Run()

	if err = app.uiManager.ui.SetRoot(app.uiManager.pages, true).
		SetFocus(app.uiManager.pages).Run(); err != nil {
		panic(err)
	}
}

func initLogger(logFilePath string) (*slog.Logger, error) {
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("can't open file: %w", err)
	}

	//nolint:exhaustruct // default logger
	handler := slog.NewTextHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler)

	slog.SetDefault(logger)

	log.SetOutput(file)
	return logger, nil
}
