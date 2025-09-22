# Stage 1: Build the application
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# CGO_ENABLED=0 is important for a static binary
# -o /user-service specifies the output file name
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /user-service .

# Stage 2: Create the final, minimal image
FROM alpine:latest

# Add ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary from the builder stage
COPY --from=builder /user-service .

# Expose port 8081 to the outside world
EXPOSE 8081

# Command to run the executable
CMD ["./user-service"]

