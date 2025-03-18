.PHONY: build run test clean frontend backend

# 默认目标
all: build

# 构建整个项目
build: backend frontend

# 构建后端
backend:
	@echo "构建后端..."
	cd backend && go build -o ../bin/meeting-mm

# 构建前端
frontend:
	@echo "构建前端..."
	cd frontend && npm run build

# 运行后端
run-backend:
	@echo "运行后端..."
	cd backend && go run main.go

# 运行前端开发服务器
run-frontend:
	@echo "运行前端开发服务器..."
	cd frontend && npm start

# 运行测试
test:
	@echo "运行测试..."
	cd backend && go test ./... -v

# 清理构建产物
clean:
	@echo "清理构建产物..."
	rm -rf bin
	rm -rf frontend/build

# 安装依赖
deps:
	@echo "安装后端依赖..."
	cd backend && go mod tidy
	@echo "安装前端依赖..."
	cd frontend && npm install

# 下载Whisper模型
download-whisper-model:
	@echo "下载Whisper模型..."
	mkdir -p whisper/models
	cd whisper && curl -L -o models/ggml-base.bin https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin

# 帮助信息
help:
	@echo "可用命令:"
	@echo "  make build          - 构建整个项目"
	@echo "  make backend        - 构建后端"
	@echo "  make frontend       - 构建前端"
	@echo "  make run-backend    - 运行后端"
	@echo "  make run-frontend   - 运行前端开发服务器"
	@echo "  make test           - 运行测试"
	@echo "  make clean          - 清理构建产物"
	@echo "  make deps           - 安装依赖"
	@echo "  make download-whisper-model - 下载Whisper模型" 