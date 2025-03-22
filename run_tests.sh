#!/bin/bash

# 会议纪要系统测试启动脚本

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # 无颜色

# 打印带颜色的消息
print_info() {
  echo -e "${YELLOW}[INFO]${NC} $1"
}

print_success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
  echo -e "${RED}[ERROR]${NC} $1"
}

print_menu() {
  echo -e "${BLUE}========== 会议纪要系统测试菜单 ==========${NC}"
  echo "1. 运行集成测试"
  echo "2. 测试Notion API连接"
  echo "3. 测试Notion同步功能"
  echo "4. 测试会议分析API"
  echo "5. 退出"
  echo -e "${BLUE}=========================================${NC}"
}

# 确保测试脚本有执行权限
chmod +x test/scripts/*.sh

# 主菜单
while true; do
  print_menu
  read -p "请选择测试类型 (1-5): " choice
  
  case $choice in
    1)
      print_info "运行集成测试..."
      ./test/scripts/integration_test.sh
      ;;
    2)
      print_info "测试Notion API连接..."
      ./test/scripts/test_notion_connection.sh
      ;;
    3)
      print_info "测试Notion同步功能..."
      ./test/scripts/notion_test.sh
      ;;
    4)
      print_info "测试会议分析API..."
      if [ ! -f "backend/meeting-mm-server" ]; then
        print_error "后端服务可执行文件不存在，请先构建后端"
        continue
      fi
      
      # 检查服务是否运行
      if ! pgrep -f "meeting-mm-server" > /dev/null; then
        print_info "后端服务未运行，正在启动..."
        cd backend && ./meeting-mm-server > backend_log.txt 2>&1 &
        sleep 3
        cd ..
      fi
      
      # 调用API
      curl -s -X POST http://localhost:8080/api/meetings/analyze \
        -H "Content-Type: application/json" \
        -d @test/json/test_analyze.json | jq .
      ;;
    5)
      print_info "退出测试菜单"
      exit 0
      ;;
    *)
      print_error "无效选择，请输入1-5之间的数字"
      ;;
  esac
  
  echo ""
  read -p "按回车键继续..."
  clear
done 