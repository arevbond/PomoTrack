package main

import (
	"embed"
	"log/slog"

	"github.com/arevbond/PomoTrack/config"
)

//go:embed assets
var f embed.FS

type Application struct {
	uiManager *UIManager
	logger    *slog.Logger
}

func NewApplication(logger *slog.Logger, cfg *config.Config) *Application {
	storage, err := newStorage("database.db")
	if err != nil {
		panic(err)
	}

	stateTaskChan := make(chan StateChangeEvent)

	app := &Application{
		logger:    logger,
		uiManager: NewUIManager(logger, cfg, stateTaskChan, NewTaskManager(logger, storage, stateTaskChan)),
	}

	app.uiManager.DefaultTimerPages()

	go app.uiManager.HandleStatesAndKeyboard()
	go app.uiManager.taskManager.HandleTaskStateChanges()

	return app
}
