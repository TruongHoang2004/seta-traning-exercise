# ===== STAGE 1: BUILD =====
FROM golang:1.23-alpine AS builder

# Cài git nếu go.mod có require từ GitHub
RUN apk add --no-cache git

WORKDIR /app

# Copy file go.mod và go.sum trước để cache dependency
COPY go.mod go.sum ./
RUN go mod download

# Copy toàn bộ source code
COPY . .

# Build binary, bỏ debug symbols để giảm size
RUN go build -ldflags="-s -w" -o auth-service ./cmd/server.go


# ===== STAGE 2: RUN =====
FROM alpine:3.19

# Cài chứng chỉ CA để gọi HTTPS
# RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary từ stage build
COPY --from=builder /app/auth-service .

# Mở port (tùy app)
EXPOSE 8080

# Chạy app
CMD ["./auth-service"]
