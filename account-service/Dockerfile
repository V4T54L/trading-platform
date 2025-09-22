# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# CGO_ENABLED=0 is important for a static binary
# -o /app/server builds the binary into the /app directory with the name 'server'
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/server ./cmd/server

# Stage 2: Create the final, minimal image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/server .

# Expose port 8084 to the outside world
EXPOSE 8084

# Command to run the executable
CMD ["./server"]

