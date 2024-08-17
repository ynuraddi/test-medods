export SECRET_KEY=124
export PG_DSN=postgresql://user:1234@localhost:5432/medods?sslmode=disable
export PG_MIGRATION_URL=file://migrations
export SMTP_HOST=
export SMTP_PORT=1025
export SMTP_USER=
export SMTP_PASS=
export SMTP_FROM=mock@gmail.com


test:
	golangci-lint run ./...
	go test ./... -v -cover

run_local:
	go run ./cmd/main.go

mock:
	mockgen -source=./internal/repository/manager.go -destination=./internal/repository/mock/mock.go
	mockgen -source=./internal/service/auth/auth.go -destination=./internal/service/auth/mock/mock.go
	mockgen -source=./internal/service/jwt/jwt.go -destination=./internal/service/jwt/mock/mock.go
	mockgen -source=./internal/service/session/session.go -destination=./internal/service/session/mock/mock.go
	mockgen -source=./internal/service/user/user.go -destination=./internal/service/user/mock/mock.go
	mockgen -source=./pkg/logger/logger.go -destination=./pkg/logger/mock/mock.go
	mockgen -source=./pkg/smtp/smtp.go -destination=./pkg/smtp/mock/mock.go

mockmail:
	docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog

run_docker:
	docker-compose build && docker-compose up
