# Build Stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod tidy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o pong-app .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/pong-app .
CMD ["./pong-app"]