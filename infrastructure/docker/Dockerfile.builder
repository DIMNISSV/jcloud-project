# infrastructure/docker/Dockerfile.builder
# Этот Dockerfile создает "сборочный стенд" со всеми зависимостями в кеше Go.

FROM golang:1.25-alpine AS builder

WORKDIR /app

# 1. Копируем ТОЛЬКО файлы модулей для кеширования зависимостей
# Этот слой будет переиспользоваться, пока зависимости не изменятся.
COPY go.work go.work.sum ./
COPY libs/go-common/go.mod ./libs/go-common/
COPY services/user-service/go.mod services/user-service/go.sum ./services/user-service/
COPY services/billing-service/go.mod services/billing-service/go.sum ./services/billing-service/
COPY services/video-service/go.mod services/video-service/go.sum ./services/video-service/

# 2. Скачиваем все зависимости для всех модулей в воркспейсе в кеш
# Это стандартный и надежный способ.
RUN go mod download

# 3. Копируем весь исходный код проекта
COPY . .