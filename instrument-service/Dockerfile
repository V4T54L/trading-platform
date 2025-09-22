# Stage 1: Build the application
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /instrument-service ./cmd/server

# Stage 2: Create the final image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /instrument-service .

EXPOSE 8082

CMD ["./instrument-service"]

