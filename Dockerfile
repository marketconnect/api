FROM golang:latest

RUN go version

ENV GOPATH=/

COPY ./ ./

RUN go mod tidy
RUN go build -o api ./cmd/main.go

CMD ["./api"]