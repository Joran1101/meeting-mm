#!/bin/bash

# 定义颜色输出
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}正在停止现有服务...${NC}"
ps aux | grep -E 'meeting-mm-server|rest_server' | grep -v grep
ps aux | grep -E 'meeting-mm-server|rest_server' | grep -v grep | awk '{print $2}' | xargs kill -9 2>/dev/null || true
echo -e "${GREEN}✅ 已停止旧进程${NC}"

# 切换到后端目录
cd backend || { echo -e "${RED}❌ 无法进入backend目录${NC}"; exit 1; }

echo -e "${YELLOW}编译后端服务...${NC}"
go build -o meeting-mm-server main.go
if [ $? -ne 0 ]; then
  echo -e "${RED}❌ 编译失败${NC}"
  exit 1
fi
echo -e "${GREEN}✅ 编译成功${NC}"

echo -e "${YELLOW}启动后端服务...${NC}"
nohup ./meeting-mm-server > ../logs/backend.log 2>&1 &
BACKEND_PID=$!

# 等待3秒检查服务是否成功启动
sleep 3
if kill -0 $BACKEND_PID 2>/dev/null; then
  echo -e "${GREEN}✅ 后端服务已启动 (PID: $BACKEND_PID)${NC}"
else
  echo -e "${RED}❌ 后端服务启动失败${NC}"
  exit 1
fi

echo -e "${YELLOW}编译REST服务器...${NC}"
mkdir -p cmd/rest_server/build
go build -o cmd/rest_server/build/rest_server cmd/rest_server/main.go
if [ $? -ne 0 ]; then
  echo -e "${RED}❌ REST服务器编译失败${NC}"
  exit 1
fi
echo -e "${GREEN}✅ REST服务器编译成功${NC}"

echo -e "${YELLOW}启动REST服务器...${NC}"
# 使用不同的端口，避免冲突
export PORT=8081
nohup ./cmd/rest_server/build/rest_server > ../logs/rest_server.log 2>&1 &
REST_PID=$!

# 等待3秒检查服务是否成功启动
sleep 3
if kill -0 $REST_PID 2>/dev/null; then
  echo -e "${GREEN}✅ REST服务器已启动 (PID: $REST_PID)${NC}"
else
  echo -e "${RED}❌ REST服务器启动失败${NC}"
  exit 1
fi

echo -e "${GREEN}所有服务已成功启动!${NC}"
echo -e "${YELLOW}Fiber API服务: ${GREEN}http://localhost:8080/api/health${NC}"
echo -e "${YELLOW}REST API服务: ${GREEN}http://localhost:8081/api/health${NC}"

echo -e "${YELLOW}服务日志:${NC}"
echo -e "${YELLOW}后端日志: ${NC}tail -f logs/backend.log"
echo -e "${YELLOW}REST服务器日志: ${NC}tail -f logs/rest_server.log" 