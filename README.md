# 会议纪要自动生成器

一个使用DeepSeek V3和Whisper进行会议纪要自动生成的工具，可以从会议录音中提取行动项和决策点。

## 功能特点

- 🎙️ 会议录音转文字（使用Whisper）
- 🤖 使用DeepSeek V3提取行动项和决策点
- 📝 生成结构化的会议纪要
- 📊 支持与Notion集成

## 技术栈

- **后端**: Go + Gin
- **前端**: React + TypeScript
- **AI模型**: DeepSeek V3 + Whisper
- **数据存储**: Notion API (可选)

## 开发环境要求

- Go 1.18+
- Node.js 16+
- npm 8+
- DeepSeek API密钥
- Notion API密钥 (可选)
- Whisper模型 (可选，支持本地运行)

## 快速开始

### 安装

1. 克隆仓库
```bash
git clone https://github.com/yourusername/meeting-mm.git
cd meeting-mm
```

2. 配置环境变量
```bash
cp backend/.env.example backend/.env
# 编辑.env文件，填入你的API密钥
```

3. 安装依赖
```bash
# 后端
cd backend
go mod download

# 前端
cd ../frontend
npm install
```

### 运行

1. 启动后端服务
```bash
cd backend
go run main.go
```

2. 启动前端服务
```bash
cd frontend
npm start
```

3. 访问应用
```
http://localhost:3000
```

## 测试

项目包含多种测试方式，确保功能正常工作。

### 运行集成测试

使用集成测试脚本可以自动测试前后端集成：

```bash
chmod +x test_integration.sh
./test_integration.sh
```

### 后端单元测试

```bash
cd backend
go test ./...
```

### 前端测试

```bash
cd frontend
npm test
```

### API测试

前端包含API测试工具，可以在浏览器中直接测试API连接、音频上传和转录分析功能。

## 项目结构

```
meeting-mm/
├── backend/             # Go后端
│   ├── api/             # API处理器和路由
│   ├── config/          # 配置管理
│   ├── services/        # 业务逻辑服务
│   └── test/            # 测试文件
├── frontend/            # React前端
│   ├── public/          # 静态资源
│   └── src/             # 源代码
│       ├── components/  # React组件
│       ├── pages/       # 页面组件
│       ├── services/    # API服务
│       ├── models/      # 数据模型
│       └── utils/       # 工具函数
└── whisper/             # Whisper模型和配置
    └── models/          # 预训练模型存放目录
```

## 贡献

欢迎提交问题和拉取请求！

## 许可证

MIT 