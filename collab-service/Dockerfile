FROM golang:1.23

WORKDIR /app

COPY . .

RUN go build -o collab-service ./cmd/main.go

CMD ["./collab-service"]
