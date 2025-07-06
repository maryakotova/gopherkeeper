# ==== Конфигурация ====
APP_NAME=gophkeeper
VERSION=$(shell git describe --tags --always --dirty || echo dev)
DATE=$(shell date +%Y-%m-%d)
MODULE_PATH=github.com/maryakotova/gopherkeeper/client
LDFLAGS=-X '$(MODULE_PATH)/cmd.buildVersion=$(VERSION)' -X '$(MODULE_PATH)/cmd.buildDate=$(DATE)'
DIST_DIR=dist

# ==== Цели ====
default: all

# Сборка под все платформы
all: clean build-linux build-windows build-macos

# Запуск (локальный)
run:
	go run ./main.go

# Сборка под Linux
build-linux:
	@echo "Сборка Gopherkeeper для Linux"
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-linux .
	@echo "Клиент Gopherkeeper для Linux собран"

# Сборка под Windows
build-windows:
	@echo "Сборка Gopherkeeper для Windows"
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-windows.exe .
	@echo "Клиент Gopherkeeper для Windows собран"

# Сборка под macOS
build-macos:
	@echo "Сборка Gopherkeeper для macOS"
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-mac .
	@echo "Клиент Gopherkeeper для macOS собран"

# Очистка
clean:
	@echo "Удаление предыдущих сборок"
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)
	@echo "Удалено"

# Версия
version:
	@echo Version: $(VERSION)
	@echo Build Date: $(DATE)