# PomoTrack 🍅

[![Go Report Card](https://goreportcard.com/badge/github.com/arevbond/PomoTrack)](https://goreportcard.com/report/github.com/arevbond/PomoTrack)

PomoTrack — это утилита для трекинга интервалов коцентрации и отдыха по методу Pomodoro. Вдохновлено [Pomofocus](https://pomofocus.io/).

![Demo](assets/demo.gif)

## Features
- Конфигурируемый таймер реалього времени;
- Добавление списка задач с необходимым количеством Pomodoros;
- Общая ститистка в виде неделього графика;
- Детальная статистка по конкретным сессиям.

## Installation

#### From source
```bash
$ git clone https://github.com/arevbond/PomoTrack
$ cd PomoTrack
$ make build
```

### Go install
```bash
go install https://github.com/arevbond/PomoTrack@latest
```
Установит бинарный файл в `$GOPATH/bin`.

## Application options
```
      --focus-duration       setup pomodoro focus intreval (default 25m)
      --break-duration       setup break interval (default 5m)
      --hidden-focus-time    hide focus clock (default false)
```
Продолжительность можно указывать в минутах (`m`) или часах (`h`), например: `25m` или `1h`.

