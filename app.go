package main

import (
	"embed"
	"log/slog"

	"github.com/arevbond/PomoTrack/config"
)

//go:embed sounds
var f embed.FS

type Application struct {
	uiManager *UIManager
	logger    *slog.Logger
}

func NewApplication(logger *slog.Logger, cfg *config.Config) *Application {
	database, err := NewStorage("database.db")
	if err != nil {
		panic(err)
	}

	stateTaskChan := make(chan StateChangeEvent)

	app := &Application{
		logger:    logger,
		uiManager: NewUIManager(logger, cfg, stateTaskChan, NewTaskManager(logger, database, stateTaskChan)),
	}

	app.uiManager.DefaultTimerPages()

	go app.uiManager.InitStateAndKeyboardHandling()
	go app.uiManager.taskTracker.HandleTaskStateChanges()

	return app
}
