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
	database, err := NewStorage(".pomotrack.UserSessions.db", logger)
	if err != nil {
		panic(err)
	}

	err = database.Migrate()
	if err != nil {
		panic(err)
	}

	stateEvents := make(chan StateEvent)

	app := &Application{
		logger:    logger,
		uiManager: NewUIManager(logger, cfg, stateEvents, NewPomodoroManager(logger, database, stateEvents), database),
	}

	return app
}

// Run - initializaion appication background tasks.
func (app *Application) Run() {
	app.uiManager.DefaultPage()

	go app.uiManager.InitStateAndKeyboardHandling()
	go app.uiManager.pomodoroTracker.HandlePomodoroStateChanges()
}
