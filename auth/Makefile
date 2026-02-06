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
	make generate-friends-api


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

generate-friends-api:
	mkdir pkg\friends_v1
	protoc --proto_path api/friends_v1 \
	--go_out=pkg/friends_v1 --go_opt=paths=source_relative \
	--go-grpc_out=pkg/friends_v1 --go-grpc_opt=paths=source_relative \
	api\friends_v1\friends.proto

test-coverage:
	go clean -testcache
	go test ./... -coverprofile=coverage.tmp.out -covermode count -coverpkg=github.com/GolZrd/micro-chat/auth/internal/service/...,github.com/GolZrd/micro-chat/auth/internal/api/...
	grep -v 'mocks\|config' coverage.tmp.out  > coverage.out
	rm coverage.tmp.out
	go tool cover -html=coverage.out;
	go tool cover -func=./coverage.out | grep "total";
	grep -sqFx "/coverage.out" .gitignore || echo "/coverage.out" >> .gitignore