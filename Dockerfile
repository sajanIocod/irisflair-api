# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git (needed for go mod download)
RUN apk add --no-cache git

# Copy go mod file first
COPY go.mod ./

# Download dependencies (generates go.sum)
RUN go mod download

# Copy source code
COPY . .

# Tidy modules
RUN go mod tidy

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Runtime stage (tiny image)
FROM alpine:3.19

WORKDIR /app

# Install CA certificates for HTTPS (MongoDB Atlas)
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/server .

# Expose port
EXPOSE 8080

# Run
CMD ["./server"]
