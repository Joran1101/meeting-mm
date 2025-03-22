#!/bin/bash

# 会议纪要自动生成器集成测试脚本
# 此脚本用于测试前端和后端的集成

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
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

# 检查命令是否存在
check_command() {
  if ! command -v $1 &> /dev/null; then
    print_error "$1 命令未找到，请安装后再运行此脚本"
    exit 1
  fi
}

# 检查必要的命令
check_command go
check_command curl

# 检查环境文件
if [ ! -f "backend/.env" ]; then
  print_info "未找到环境配置文件，将使用示例配置"
  if [ -f "backend/.env.example" ]; then
    cp backend/.env.example backend/.env
    print_success "已复制示例配置到 backend/.env"
  else
    print_error "示例配置文件不存在，请创建 backend/.env 文件"
    exit 1
  fi
fi

# 创建测试目录
mkdir -p test/testdata

# 启动后端服务
print_info "正在启动后端服务..."
cd backend
go build -o meeting-mm-server
if [ $? -ne 0 ]; then
  print_error "后端编译失败"
  exit 1
fi

# 在后台运行后端服务
./meeting-mm-server &
BACKEND_PID=$!
cd ..

# 等待后端服务启动
print_info "等待后端服务启动..."
sleep 3

# 测试健康检查API
print_info "测试健康检查API..."
HEALTH_RESPONSE=$(curl -s http://localhost:8080/api/health)
if [ $? -ne 0 ]; then
  print_error "无法连接到后端服务"
  kill $BACKEND_PID
  exit 1
fi

if command -v jq &> /dev/null; then
  HEALTH_STATUS=$(echo $HEALTH_RESPONSE | jq -r '.status')
  if [ "$HEALTH_STATUS" == "ok" ]; then
    print_success "健康检查API测试通过"
  else
    print_error "健康检查API测试失败"
    kill $BACKEND_PID
    exit 1
  fi
else
  # 如果没有jq，使用简单的字符串匹配
  if [[ $HEALTH_RESPONSE == *"\"status\":\"ok\""* ]]; then
    print_success "健康检查API测试通过"
  else
    print_error "健康检查API测试失败"
    kill $BACKEND_PID
    exit 1
  fi
fi

# 检查是否有DeepSeek API密钥
DEEPSEEK_API_KEY=$(grep "DEEPSEEK_API_KEY" backend/.env | cut -d '=' -f2)
if [ -z "$DEEPSEEK_API_KEY" ] || [ "$DEEPSEEK_API_KEY" == "your_deepseek_api_key_here" ]; then
  print_info "跳过转录分析API测试，因为没有配置DeepSeek API密钥"
else
  # 测试转录分析API
  print_info "测试转录分析API..."
  TRANSCRIPT_TEST_DATA='{
    "title": "测试会议",
    "transcript": "张三：大家好，今天我们讨论项目进度。\n李四：我已经完成了前端开发，需要王五测试一下。\n王五：好的，我明天会进行测试。\n张三：那我们决定下周一发布第一个版本。"
  }'

  ANALYZE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/meetings/analyze \
    -H "Content-Type: application/json" \
    -d "$TRANSCRIPT_TEST_DATA")

  if [ $? -ne 0 ]; then
    print_error "转录分析API请求失败"
    kill $BACKEND_PID
    exit 1
  fi

  # 检查响应是否包含会议数据
  if command -v jq &> /dev/null; then
    if echo $ANALYZE_RESPONSE | jq -e '.meeting' > /dev/null; then
      print_success "转录分析API测试通过"
    else
      print_error "转录分析API测试失败"
      echo $ANALYZE_RESPONSE
      kill $BACKEND_PID
      exit 1
    fi
  else
    # 如果没有jq，使用简单的字符串匹配
    if [[ $ANALYZE_RESPONSE == *"\"meeting\""* ]]; then
      print_success "转录分析API测试通过"
    else
      print_error "转录分析API测试失败"
      echo $ANALYZE_RESPONSE
      kill $BACKEND_PID
      exit 1
    fi
  fi
fi

# 如果前端目录存在package.json，尝试安装依赖并启动前端
if [ -f "frontend/package.json" ]; then
  print_info "检测到前端项目，尝试安装依赖..."
  cd frontend
  
  # 检查npm是否存在
  if command -v npm &> /dev/null; then
    npm install
    if [ $? -ne 0 ]; then
      print_error "前端依赖安装失败"
    else
      print_success "前端依赖安装成功"
      
      # 启动前端服务（可选）
      print_info "可以使用以下命令启动前端开发服务器："
      echo "cd frontend && npm start"
    fi
  else
    print_info "未检测到npm，跳过前端依赖安装"
  fi
  
  cd ..
fi

# 测试完成，关闭后端服务
print_info "正在关闭后端服务..."
kill $BACKEND_PID

print_success "集成测试完成"
print_info "如需进一步测试，请运行以下命令："
echo "1. 启动后端: cd backend && ./meeting-mm-server"
echo "2. 启动前端: cd frontend && npm start (如果已安装npm)"
echo "3. 访问前端: http://localhost:3000" 