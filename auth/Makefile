ROOT_DIR := ..
include $(ROOT_DIR)/.env

local-migration-status:
	goose -dir ./migrations postgres ${PG_AUTH_DSN} status -v

local-migration-up:
	goose -dir ./migrations postgres ${PG_AUTH_DSN} up -v

local-migration-down:
	goose -dir ./migrations postgres ${PG_AUTH_DSN} down -v


generate:
	make generate-access-api
	make generate-auth-api
	make generate-user-api

generate-access-api:
	mkdir pkg\access_v1
	protoc --proto_path api/access_v1 \
	--go_out=pkg/access_v1 --go_opt=paths=source_relative \
	--go-grpc_out=pkg/access_v1 --go-grpc_opt=paths=source_relative \
	api\access_v1\access.proto

generate-auth-api:
	mkdir pkg\auth_v1
	protoc --proto_path api/auth_v1 \
	--go_out=pkg/auth_v1 --go_opt=paths=source_relative \
	--go-grpc_out=pkg/auth_v1 --go-grpc_opt=paths=source_relative \
	api\auth_v1\auth.proto

generate-user-api:
	mkdir pkg\user_v1
	protoc --proto_path api/user_v1 \
	--go_out=pkg/user_v1 --go_opt=paths=source_relative \
	--go-grpc_out=pkg/user_v1 --go-grpc_opt=paths=source_relative \
	api\user_v1\user.proto
