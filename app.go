package main

import (
	"embed"
	"log"
	"log/slog"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

//go:embed assets
var f embed.FS

type Application struct {
	uiManager *UIManager
	logger    *slog.Logger
}

func NewApplication(logger *slog.Logger) *Application {
	database, err := NewStorage("database.db")
	if err != nil {
		panic(err)
	}

	stateTaskChan := make(chan StateChangeEvent)

	app := &Application{
		logger:    logger,
		uiManager: NewUIManager(logger, stateTaskChan, NewTaskManager(logger, database, stateTaskChan)),
	}

	app.uiManager.DefaultTimerPages()

	go app.uiManager.HandleStatesAndKeyboard()
	go app.uiManager.taskManager.HandleTaskStateChanges()

	return app
}

func playClickSound() {
	file, err := f.Open("assets/click-sound2.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	_ = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}

func playEndSound() {
	file, err := f.Open("assets/end-sound.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	_ = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
