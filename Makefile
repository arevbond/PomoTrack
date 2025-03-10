APP_NAME = pomotrack
BUILD_DIR = build
GO_FILES = $(shell find . -name '*.go')

run: build
	./$(APP_NAME)

test:
	@go test ./...

deps:
	go mod tidy
	go mod vendor

fmt:
	go fmt ./...

# Сборка бинарного файла для текущей платформы
build:
	go build -o $(APP_NAME)

# Кросс-компиляция для всех поддерживаемых платформ
build-all:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe

# Обновление go.sum файла
verify:
	go mod verify

# Запуск тестов с покрытием
coverage:
	go test -coverprofile=c.out ./...

migrations:
	@goose -dir=./migrations sqlite3 ./database.db up

count-lines:
	@grep -r -h . *.go | wc -l