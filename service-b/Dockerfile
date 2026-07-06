FROM golang:1.26.1-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o service-b ./cmd/api/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/service-b .
ENTRYPOINT ["./service-b"]
