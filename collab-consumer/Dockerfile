# ===== STAGE 1: BUILD =====
FROM golang:1.23-alpine AS builder

# Cài git nếu go.mod có require từ GitHub
RUN apk add --no-cache git

WORKDIR /app

# Copy và tải dependencies trước để cache
COPY go.mod go.sum ./
RUN go mod download

# Copy toàn bộ source code vào
COPY . .

# Build binary, bỏ debug symbols để giảm size
RUN go build -ldflags="-s -w" -o collab-consumer ./cmd/main.go


# ===== STAGE 2: RUN =====
FROM alpine:3.19

# # Cài chứng chỉ CA để gọi HTTPS
# RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary từ stage build
COPY --from=builder /app/collab-consumer .

# Chạy app
CMD ["./collab-consumer"]
