FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o /app/main ./app/cmd/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main ./main

EXPOSE 8080
CMD ["./main"]
