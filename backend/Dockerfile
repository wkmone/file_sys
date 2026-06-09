FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev

# ========== 设置国内Go代理，解决超时 ==========
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/server .

RUN mkdir -p /app/storage

EXPOSE 8080
ENTRYPOINT ["/app/server"]
