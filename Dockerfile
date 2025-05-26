FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.* .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/main.go

FROM alpine:3.20

RUN apk add --no-cache tzdata

COPY --from=builder /app/main /app/oncall_notifier

CMD ["/app/oncall_notifier", "-config", "/app/config.yml"]
