.PHONY: build test lint-self install-check

# Сборка плагина
build:
	CGO_ENABLED=1 go build -buildmode=plugin -o loglinter.so ./plugin/main.go

# Запуск юнит-тестов
test:
	go test -v ./pkg/analyzer/...

# Пример запуска на тестовых данных
check: build
	golangci-lint run ./pkg/analyzer/testdata/src/a/...

# Полный цикл для проверки
all: test check