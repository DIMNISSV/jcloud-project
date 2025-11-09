ARG BUILDER_IMAGE=golang:1.25-alpine
# Этап 1: Используем наш готовый сборочный стенд
FROM ${BUILDER_IMAGE} AS builder

# Этап 2: Собираем конкретный бинарник ИЗ КОРНЯ /app
# Мы не меняем WORKDIR, чтобы команда видела /app/vendor
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-service ./services/user-service/cmd/api

# Этап 3: Финальный образ
FROM scratch
WORKDIR /
COPY --from=builder /user-service /user-service
EXPOSE 8080
ENTRYPOINT ["/user-service"]