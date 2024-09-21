FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /chat-server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /chat-server .
COPY --from=builder /app/pkg/chat/v1/chat.swagger.json /root/pkg/chat/v1/chat.swagger.json

EXPOSE 50051 8080

ENTRYPOINT ["./chat-server"]
CMD []