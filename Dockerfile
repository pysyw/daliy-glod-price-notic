FROM golang:1.24-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
# 明确指定目标平台为 amd64，确保在云平台上能正常运行
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gold-price-notic .

# 运行镜像
FROM alpine:latest

WORKDIR /app

# 安装证书（HTTPS请求需要）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为中国
ENV TZ=Asia/Shanghai

COPY --from=builder /app/gold-price-notic .

CMD ["./gold-price-notic"]
