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
FROM mcr.microsoft.com/playwright:v1.41.1-jammy

WORKDIR /app

# 从builder阶段复制编译好的程序
COPY --from=builder /app/main .

# 复制配置文件和资源
COPY hash_store.json .
COPY update.txt .
COPY utils/california ./utils/california

# 添加必要的系统依赖
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# 设置时区为上海
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# 运行应用
CMD ["./main"]