FROM golang:1.24-alpine as builder

WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 创建最终的精简镜像
FROM mcr.microsoft.com/playwright:v1.52.0-jammy

WORKDIR /app

# 从builder阶段复制编译好的程序
COPY --from=builder /app/main .

# 复制配置文件和资源
COPY hash_store.json .
COPY update.txt .
COPY .env /app/.env

# 运行应用
CMD ["./main"]