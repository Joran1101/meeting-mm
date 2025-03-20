#!/bin/bash

# 会议纪要应用启动脚本
# 此脚本用于同时启动前端和后端服务

# 彩色输出函数
print_green() {
  echo -e "\033[0;32m$1\033[0m"
}

print_yellow() {
  echo -e "\033[0;33m$1\033[0m"
}

print_red() {
  echo -e "\033[0;31m$1\033[0m"
}

# 项目根目录
PROJECT_ROOT="$(pwd)"

# 检查是否有上次启动的残留进程
print_yellow "正在检查残留进程..."
ps aux | grep -E 'meeting-mm-server|react-scripts' | grep -v grep

# 询问是否终止现有进程
read -p "是否终止已存在的进程? (y/n): " kill_processes
if [[ "$kill_processes" == "y" || "$kill_processes" == "Y" ]]; then
  print_yellow "正在终止现有进程..."
  ps aux | grep -E 'meeting-mm-server|react-scripts' | grep -v grep | awk '{print $2}' | xargs kill -9 2>/dev/null || true
  print_green "✅ 进程已终止"
fi

# 创建日志目录
mkdir -p "$PROJECT_ROOT/logs"

# 启动后端服务
start_backend() {
  print_yellow "启动后端服务..."
  cd "$PROJECT_ROOT/backend" || exit 1
  
  # 检查可执行文件是否存在
  if [[ ! -f "./meeting-mm-server" ]]; then
    print_yellow "后端可执行文件不存在，尝试编译..."
    go build -o meeting-mm-server main.go
    if [[ $? -ne 0 ]]; then
      print_red "❌ 后端编译失败！"
      return 1
    fi
    print_green "✅ 后端编译成功"
  fi
  
  # 启动后端，将输出重定向到日志文件
  ./meeting-mm-server > "$PROJECT_ROOT/logs/backend_$(date +%Y%m%d).log" 2>&1 &
  BACKEND_PID=$!
  
  # 等待3秒，确认后端启动成功
  sleep 3
  if kill -0 $BACKEND_PID 2>/dev/null; then
    print_green "✅ 后端服务已启动 (PID: $BACKEND_PID)"
  else
    print_red "❌ 后端服务启动失败"
    return 1
  fi
  
  return 0
}

# 启动前端服务
start_frontend() {
  print_yellow "启动前端服务..."
  cd "$PROJECT_ROOT/frontend" || exit 1
  
  # 检查node_modules是否存在
  if [[ ! -d "node_modules" ]]; then
    print_yellow "未找到node_modules，需要先安装依赖..."
    npm install
    if [[ $? -ne 0 ]]; then
      print_red "❌ 前端依赖安装失败！"
      return 1
    fi
    print_green "✅ 前端依赖安装成功"
  fi
  
  # 启动前端，将输出重定向到日志文件，使用不同端口避免冲突
  PORT=3001 npm start > "$PROJECT_ROOT/logs/frontend_$(date +%Y%m%d).log" 2>&1 &
  FRONTEND_PID=$!
  
  # 等待5秒，确认前端启动成功
  sleep 5
  if kill -0 $FRONTEND_PID 2>/dev/null; then
    print_green "✅ 前端服务已启动 (PID: $FRONTEND_PID)"
  else
    print_red "❌ 前端服务启动失败"
    return 1
  fi
  
  return 0
}

# 主函数
main() {
  print_green "=== 会议纪要应用启动工具 ==="
  
  # 启动后端
  start_backend
  if [[ $? -ne 0 ]]; then
    print_red "启动过程中止"
    exit 1
  fi
  
  # 启动前端
  start_frontend
  if [[ $? -ne 0 ]]; then
    print_red "启动过程中止"
    exit 1
  fi
  
  print_green "=== 所有服务已成功启动 ==="
  print_green "后端服务运行在: http://localhost:8080"
  print_green "前端服务运行在: http://localhost:3001"
  print_yellow "日志文件保存在: $PROJECT_ROOT/logs/"
  print_yellow "按 Ctrl+C 结束此会话不会终止服务。如需停止服务，请使用 stop.sh 脚本或手动终止进程。"
}

# 运行主函数
main 