FROM golang:1.23

WORKDIR /app

COPY . .

RUN go build -o auth-service ./cmd/server.go

CMD ["./auth-service"]
