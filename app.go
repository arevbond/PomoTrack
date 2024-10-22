package main

import (
	"embed"
	"github.com/arevbond/PomoTrack/config"
	"log/slog"
)

//go:embed assets
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

	go app.uiManager.HandleStatesAndKeyboard()
	go app.uiManager.taskManager.HandleTaskStateChanges()

	return app
}
