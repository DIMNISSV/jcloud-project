# services/billing-service/Dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /billing-service ./cmd/api

FROM scratch
WORKDIR /
COPY --from=builder /billing-service /billing-service
EXPOSE 8082
ENTRYPOINT ["/billing-service"]