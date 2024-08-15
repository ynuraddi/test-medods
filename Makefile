export SECRET_KEY=124
export PG_DSN=postgresql://user:1234@localhost:5432/medods?sslmode=disable
export PG_MIGRATION_URL=file://migrations

test:
	golangci-lint run ./...
	go test ./... -v -cover

run:
	go run ./cmd/main.go

mock:
	mockgen -source=./internal/repository/manager.go -destination=./internal/repository/mock/mock.go
	mockgen -source=./internal/service/auth/auth.go -destination=./internal/service/auth/mock/mock.go
	mockgen -source=./internal/service/jwt/jwt.go -destination=./internal/service/jwt/mock/mock.go
	mockgen -source=./internal/service/session/session.go -destination=./internal/service/session/mock/mock.go
	mockgen -source=./internal/service/user/user.go -destination=./internal/service/user/mock/mock.go
	mockgen -source=./pkg/logger/logger.go -destination=./pkg/logger/mock/mock.go
