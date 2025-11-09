ARG BUILDER_IMAGE=golang:1.25-alpine
# Этап 1: Используем наш готовый сборочный стенд
FROM ${BUILDER_IMAGE} AS builder

# Этап 2: Собираем конкретный бинарник ИЗ КОРНЯ /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /video-service ./services/video-service/cmd/api

# Этап 3: Финальный образ
FROM scratch
WORKDIR /
COPY --from=builder /video-service /video-service
EXPOSE 8081
ENTRYPOINT ["/video-service"]