ROOT_DIR := ..
include $(ROOT_DIR)/.env

local-migration-status:
	goose -dir ./migrations postgres ${PG_CHAT_DSN} status -v

local-migration-up:
	goose -dir ./migrations postgres ${PG_CHAT_DSN} up -v

local-migration-down:
	goose -dir ./migrations postgres ${PG_CHAT_DSN} down -v


generate:
	make generate-chat-api

generate-chat-api:
	mkdir pkg\chat_v1
	protoc --proto_path api/chat_v1 \
	--go_out=pkg/chat_v1 --go_opt=paths=source_relative \
	--go-grpc_out=pkg/chat_v1 --go-grpc_opt=paths=source_relative \
	api\chat_v1\chat.proto

	