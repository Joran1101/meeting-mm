# 多阶段构建

# 后端构建阶段
FROM golang:1.20-alpine AS backend-builder
WORKDIR /app

# 安装依赖
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download

# 复制源代码
COPY backend ./backend

# 构建后端
RUN cd backend && go build -o /app/bin/meeting-mm

# 前端构建阶段
FROM node:16-alpine AS frontend-builder
WORKDIR /app

# 安装依赖
COPY frontend/package.json frontend/package-lock.json* ./frontend/
RUN cd frontend && npm ci

# 复制源代码
COPY frontend ./frontend

# 构建前端
RUN cd frontend && npm run build

# 最终镜像
FROM alpine:3.16
WORKDIR /app

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 复制Whisper模型
COPY whisper/models /app/whisper/models

# 复制构建产物
COPY --from=backend-builder /app/bin/meeting-mm /app/
COPY --from=frontend-builder /app/frontend/build /app/public
COPY backend/.env.example /app/.env

# 设置环境变量
ENV PORT=8080
ENV ENV=production
ENV WHISPER_MODEL_PATH=/app/whisper/models/ggml-base.bin
ENV USE_LOCAL_WHISPER=true

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["/app/meeting-mm"] 