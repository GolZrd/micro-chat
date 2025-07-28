FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o auth cmd/grpc_server/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/auth .

EXPOSE 50051

CMD ["./auth"]