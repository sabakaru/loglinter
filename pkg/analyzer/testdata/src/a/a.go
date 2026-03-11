package a

import (
	"log/slog"
)

func invalidLogs() {
	slog.Info("Starting server on port 8080")   // want "log message must start with a lowercase letter"
	slog.Error("Failed to connect to database") // want "log message must start with a lowercase letter"

	slog.Info("запуск сервера")                    // want "log message must be in English only"
	slog.Error("ошибка подключения к базе данных") // want "log message must be in English only"

	slog.Info("server started! \U0001F680")       // want "log message must not contain emojis"
	slog.Error("connection failed!!!")            // want "log message must not contain special characters"
	slog.Warn("warning: something went wrong...") // want "log message must not contain special characters"

	password := "123"
	slog.Info("user password: " + password) // want "log message must not contain sensitive data"

	apiKey := "key"
	slog.Debug("api_key=" + apiKey) // want "log message must not contain sensitive data"

	token := "token"
	slog.Debug("token=" + token) // want "log message must not contain sensitive data"
}

func validLogs() {
	slog.Info("starting server on port 8080")
	slog.Error("failed to connect to database")
	slog.Info("starting server")
	slog.Info("server started")
	slog.Error("connection failed")
	slog.Warn("something went wrong")
	slog.Info("user authenticated successfully")
	slog.Debug("api request completed")
	slog.Info("token validated")
}
