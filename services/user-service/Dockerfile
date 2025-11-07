# services/user-service/Dockerfile

# Stage 1: Build the application
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
# CGO_ENABLED=0 is important for a static binary in a scratch image
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-service ./cmd/api

# Stage 2: Create the final, minimal image
FROM scratch

WORKDIR /
COPY --from=builder /user-service /user-service

# Expose port 8080
EXPOSE 8080

# Command to run the executable
ENTRYPOINT ["/user-service"]