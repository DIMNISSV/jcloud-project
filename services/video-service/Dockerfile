# services/video-service/Dockerfile

FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /video-service ./cmd/api

FROM scratch
WORKDIR /
COPY --from=builder /video-service /video-service
EXPOSE 8081
ENTRYPOINT ["/video-service"]