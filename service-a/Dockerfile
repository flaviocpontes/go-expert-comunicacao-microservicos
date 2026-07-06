FROM golang:1.26.1-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o service-a ./cmd/api/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/service-a .
ENTRYPOINT ["./service-a"]
