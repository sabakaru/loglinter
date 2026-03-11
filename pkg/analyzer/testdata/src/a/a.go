package a

import (
	"log/slog"
)

func invalidLogs() {
	// rule 1 - first letter must be lowercase
	slog.Info("Starting server on port 8080")   // want "log message must start with a lowercase letter"
	slog.Error("Failed to connect to database") // want "log message must start with a lowercase letter"

	// rule 2 - must be in English only
	slog.Info("запуск сервера")                    // want "log message must be in English only"
	slog.Error("ошибка подключения к базе данных") // want "log message must be in English only"

	// rule 3 - forbidden symbols or emojis
	slog.Info("server started!")          // want "log message contains forbidden symbols or emojis"
	slog.Error("connection failed!!!")    // want "log message contains forbidden symbols or emojis"
	slog.Warn("something went wrong...")  // want "log message contains forbidden symbols or emojis"
	slog.Info("we are flying \U0001F680") // want "log message contains forbidden symbols or emojis"

	// rule 4 - sensitive data
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
	slog.Info("processing item 42 of 100")
}
