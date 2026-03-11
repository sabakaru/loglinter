# Log Linter

[![Linter CI](https://github.com/sabakaru/loglinter/actions/workflows/ci.yml/badge.svg)](https://github.com/sabakaru/loglinter/actions/workflows/ci.yml)

Cпециализированный статический анализатор (линтер) для языка Go, направленный на поддержание чистоты, безопасности и единообразия лог-сообщений.

Линтер полностью совместим с `golangci-lint` в виде плагина и легко встраивается в `pre-commit` хуки благодаря наличию автономного бинарного файла.

## Поддерживаемые логгеры
* `log/slog`
* `go.uber.org/zap`

---

## Проверяемые правила

### 1. Строчная буква в начале сообщения (`check-lowercase`)
Лог-сообщения должны начинаться со строчной буквы (кроме случаев, когда первое слово является аббревиатурой или именем собственным, начинающимся с цифры).

```go
// ❌ Неправильно
log.Info("Starting server on port 8080")
slog.Error("Failed to connect to database")

// ✅ Правильно
log.Info("starting server on port 8080")
slog.Error("failed to connect to database")
```
*Поддерживает **Auto-Fix** При запуске с флагом `-fix` линтер автоматически переведёт первую букву в нижний регистр.*

### 2. Только английский язык (`check-english`)
Лог-сообщения предназначены для технической команды и могут собираться в международные логирующие системы. Ограничение символов латиницей помогает избежать проблем с кодировкой и упрощает парсинг логов.

```go
// ❌ Неправильно
log.Info("запуск сервера")
log.Error("ошибка подключения к базе данных")

// ✅ Правильно
log.Info("starting server")
log.Error("failed to connect to database")
```

### 3. Запрет спецсимволов и эмодзи (`check-specials`)
Излишняя пунктуация и эмодзи засоряют логи, затрудняя их индексацию и поиск в системах типа ELK-стека. Сообщение должно быть сухим фактом.

```go
// ❌ Неправильно
log.Info("server started! 🚀")
log.Error("connection failed!!!")
log.Warn("warning: something went wrong...\n")

// ✅ Правильно
log.Info("server started")
log.Error("connection failed")
log.Warn("something went wrong")
```

### 4. Отсутствие чувствительных данных (`check-sensitive`)
Линтер защищает от утечек конфиденциальной информации. Он анализирует как содержимое строк, так и названия переменных, передаваемых в логгер.

```go
// ❌ Неправильно
log.Info("user password:\n" + password)
log.Debug("api_key=\n" + apiKey)
log.Info("token:\n" + token)

// ✅ Правильно
log.Info("user authenticated successfully")
log.Debug("api request completed")
log.Info("token validated")
```

---

## Установка и использование

У вас есть два основных способа интеграции лог-линтера в свой проект.

### Способ 1. Как плагин для `golangci-lint` (Рекомендуемый)

1. **Сборка плагина:**
```bash
# Клонируем репозиторий:
git clone https://github.com/sabakaru/loglinter.git
cd loglinter

# Собираем файл плагина (loglinter.so) для golangci-lint
make build
```

2. **Подключение в вашем проекте:**
Разместите собранный `.so` файл в вашем проекте и добавьте конфигурацию в ваш `.golangci.yml`:

```yaml
linters:
  disable-all: true
  enable:
    - loglinter

linters-settings:
  custom:
    loglinter:
      path: ./loglinter.so # Путь до собранного плагина
      description: Анализатор чистоты и безопасности лог-сообщений
      original-env: false
```

### Способ 2. Интеграция через `pre-commit` (Zero конфигурации)

Вы можете использовать `loglinter` вообще без настройки `golangci-lint`, добавив его прямо в `pre-commit` конвейер вашего проекта. Линтер скачается, скомпилируется и запустится как самостоятельный анализатор!

Добавьте в свой `.pre-commit-config.yaml`:
```yaml
repos:
  - repo: https://github.com/sabakaru/loglinter
    rev: main
    hooks:
      - id: loglinter
```

---

## Конфигурация и гибкость по умолчанию

Все 4 главных правила включены по умолчанию.
Конфидентциальными словами по умолчанию (rule 4) считаются: `password`, `apiKey`, `api_key`, `token`.

У вас есть возможность отключать правила точечно или переопределить список критичных слов.

### Конфигурация в `golangci-lint` (.golangci.yml)
Все параметры настраиваются в блоке `settings` кастомного плагина:

```yaml
linters-settings:
  custom:
    loglinter:
      path: ./loglinter.so
      settings:
        # Кастомные слова для правила на чувствительные данные
        sensitive_words: ["password", "apiKey", "api_key", "token", "secret", "cvv"]

        # Индивидуальные тоглы для правил
        check_lowercase: true
        check_english: true
        check_specials: true
        check_sensitive: true
```

### Конфигурация в `pre-commit` или CLI
При использовании standalone бинарника (через pre-commit) конфигурация осуществляется через `args`:

```yaml
      - id: loglinter
        args:
          - -check-lowercase=false  # Отключить проверку регистра
          - -check-specials=true
          - -sensitive-words=password,token,secret,jwt
```

---

## 🚀 CI/CD и Автоматизация

Проект полностью настроен для использования в Continuous Integration pipeline'ах:
* **GitHub Actions:** В репозитории уже настроен процесс автоматической сборки плагина `.so` и прогона Unit-тестов при каждом Pull Request'e и push'е (см. [`.github/workflows/ci.yml`](.github/workflows/ci.yml)). Плагин автоматически собирается как Artifact.
* **Pre-commit:** В проекте настроена проверка кода перед любыми коммитами через `pre-commit-config.yaml` (go-fmt, go-imports, check-yaml и сам loglinter).
* **Make:** Присутствует `Makefile` с готовыми командами для разработки (`make test`, `make build`, `make check`).

---

## Структура проекта
* `pkg/analyzer/` — Основная логика работы линтера (правила, AST парсинг, тесты).
  * `testdata/a/` — Unit-тесты для анализатора (Invalid/Valid случаи).
* `cmd/loglinter/` — Standalone обёртка для работы линтера как самостоятельного CLI приложения (`singlechecker`).
* `plugin/` — Обёртка для динамически подключаемого плагина `.so` для `golangci-lint`.
