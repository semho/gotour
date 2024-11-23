FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN mkdir -p /build/bin

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/chat-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/migrate ./cmd/migrator

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/chat-server .
COPY --from=builder /app/migrate ./bin/migrate
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/pkg/chat/v1/chat.swagger.json ./pkg/chat/v1/chat.swagger.json

EXPOSE 50051 8080

ENTRYPOINT ["./chat-server"]