# Stage 1: Build the application
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /order-service ./cmd/server

# Stage 2: Create the final image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /order-service .

EXPOSE 8083

CMD ["./order-service"]

