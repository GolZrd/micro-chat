include .env

# Указываем PowerShell как shell
SHELL := powershell.exe
.SHELLFLAGS := -NoProfile -Command

LOCAL_BIN:=$(CURDIR)\bin

LOCAL_MIGRATION_DIR = $(MIGRATION_DIR)
LOCAL_MIGRATION_DSN="host=localhost port=${PG_PORT} dbname=${PG_DATABASE_NAME} user=${PG_USER} password=${PG_PASSWORD} sslmode=disable"


local-migration-status:
	goose -dir $(LOCAL_MIGRATION_DIR) postgres ${LOCAL_MIGRATION_DSN} status -v

local-migration-up:
	$(GOOSE) -dir $(LOCAL_MIGRATION_DIR) postgres ${LOCAL_MIGRATION_DSN} up -v

local-migration-down:
	$(GOOSE) -dir $(LOCAL_MIGRATION_DIR) postgres ${LOCAL_MIGRATION_DSN} down -v


generate:
	make generate-chat-api

generate-chat-api:
	mkdir pkg\chat_v1
	protoc --proto_path api/chat_v1 \
	--go_out=pkg/chat_v1 --go_opt=paths=source_relative \
	--go-grpc_out=pkg/chat_v1 --go-grpc_opt=paths=source_relative \
	api\chat_v1\chat.proto