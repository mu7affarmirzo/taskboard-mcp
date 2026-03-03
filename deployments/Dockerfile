FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bot ./cmd/bot

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata wget
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /app/bot .
RUN mkdir -p /app/data && chown appuser:appuser /app/data
USER appuser
ENV DATABASE_PATH=/app/data/bot.db
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/healthz || exit 1
CMD ["./bot"]
