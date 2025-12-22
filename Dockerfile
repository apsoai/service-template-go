# Go/Gin Service Dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for go mod download
RUN apk add --no-cache git

# Install swag CLI for Swagger docs generation
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy all source code first (needed for go mod tidy to analyze imports)
COPY . .

# Download dependencies (tidy first to generate go.sum if needed)
RUN go mod tidy && go mod download

# Generate Swagger documentation
RUN swag init -g cmd/main.go -o docs

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Production stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS and curl for healthcheck
RUN apk --no-cache add ca-certificates curl

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 3000

# Start the application
CMD ["./main"]
