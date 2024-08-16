FROM golang:1.21.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o medods ./cmd/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/medods .
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/config/ /app/config

LABEL maintainers = "ynuraddi"
LABEL version = "1.0"

EXPOSE 8080

CMD ["./medods"]