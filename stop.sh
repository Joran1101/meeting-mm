#!/bin/bash

# 会议纪要应用停止脚本
# 此脚本用于停止前端和后端服务

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

# 主函数
main() {
  print_yellow "=== 停止会议纪要应用服务 ==="
  
  # 查找后端服务进程
  print_yellow "查找后端服务进程..."
  BACKEND_PROCESSES=$(ps aux | grep 'meeting-mm-server' | grep -v grep | awk '{print $2}')
  
  if [[ -n "$BACKEND_PROCESSES" ]]; then
    print_yellow "找到后端进程: $BACKEND_PROCESSES"
    print_yellow "正在停止后端服务..."
    echo "$BACKEND_PROCESSES" | xargs kill -9 2>/dev/null
    if [[ $? -eq 0 ]]; then
      print_green "✅ 后端服务已停止"
    else
      print_red "❌ 停止后端服务失败"
    fi
  else
    print_yellow "未找到运行中的后端服务"
  fi
  
  # 查找前端服务进程
  print_yellow "查找前端服务进程..."
  FRONTEND_PROCESSES=$(ps aux | grep 'react-scripts' | grep -v grep | awk '{print $2}')
  
  if [[ -n "$FRONTEND_PROCESSES" ]]; then
    print_yellow "找到前端进程: $FRONTEND_PROCESSES"
    print_yellow "正在停止前端服务..."
    echo "$FRONTEND_PROCESSES" | xargs kill -9 2>/dev/null
    if [[ $? -eq 0 ]]; then
      print_green "✅ 前端服务已停止"
    else
      print_red "❌ 停止前端服务失败"
    fi
  else
    print_yellow "未找到运行中的前端服务"
  fi
  
  # 查找node进程（防止有些React进程没有正确识别）
  print_yellow "检查其他相关进程..."
  NODE_PROCESSES=$(ps aux | grep 'node.*frontend' | grep -v grep | awk '{print $2}')
  
  if [[ -n "$NODE_PROCESSES" ]]; then
    print_yellow "找到其他相关进程: $NODE_PROCESSES"
    print_yellow "正在停止其他相关进程..."
    echo "$NODE_PROCESSES" | xargs kill -9 2>/dev/null
    if [[ $? -eq 0 ]]; then
      print_green "✅ 其他相关进程已停止"
    else
      print_red "❌ 停止其他相关进程失败"
    fi
  fi
  
  # 最后检查是否还有残留进程
  REMAINING=$(ps aux | grep -E 'meeting-mm-server|react-scripts|node.*frontend' | grep -v grep)
  
  if [[ -n "$REMAINING" ]]; then
    print_yellow "仍有以下相关进程在运行:"
    echo "$REMAINING"
    print_yellow "如需手动终止，请使用命令: kill -9 <PID>"
  else
    print_green "=== 所有服务已成功停止 ==="
  fi
}

# 执行主函数
main 