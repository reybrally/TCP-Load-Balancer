FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o tcp-lb cmd/main.go

FROM alpine:3.18
RUN apk --no-cache update && \
    apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/tcp-lb .
COPY config.yaml .

EXPOSE 8080 9090

CMD ["./tcp-lb"]
