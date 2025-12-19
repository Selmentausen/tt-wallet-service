FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o wallet-app ./cmd/api/main.go

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/wallet-app .
EXPOSE 8080

CMD ["./wallet-app"]