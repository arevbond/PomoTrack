package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
)

func main() {
	logger, err := initLogger("app.log")
	if err != nil {
		panic(err)
	}

	app := NewApplication(logger)

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
