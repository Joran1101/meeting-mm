# 会议纪要自动生成器

一个使用DeepSeek和Whisper进行会议纪要自动生成的工具，可以从会议录音中提取行动项和决策点。

## 功能特点

- 🎙️ 会议录音转文字（使用Whisper）
- 🤖 使用DeepSeek提取行动项和决策点
- 📝 生成结构化的会议纪要
- 📊 支持与Notion集成
- 📱 响应式Web界面，支持多种设备
- 🔄 自动同步会议记录到Notion
- 📋 支持多种日期格式

## 技术栈

- **后端**: Go + Fiber
- **前端**: React + TypeScript + Material-UI
- **AI模型**: DeepSeek Chat + Whisper
- **数据存储**: Notion API
- **测试**: 集成测试 + 单元测试

## 开发环境要求

- Go 1.18+
- Node.js 16+
- npm 8+
- Python 3.8+ (用于Whisper)
- DeepSeek API密钥
- Notion API密钥
- Whisper模型 (本地运行)

## 项目当前状态

- ✅ 音频上传和转录功能
- ✅ 会议内容分析（摘要、待办事项、决策点）
- ✅ Markdown格式报告生成
- ✅ Notion集成基础功能
- ✅ 测试脚本和文档整理
- ⏳ 说话人识别功能
- ⏳ 音频段落自动划分
- ⏳ UI/UX优化

查看完整的[待办事项列表](TODO.md)了解计划中的功能和已知问题。

## 快速开始

### 安装

1. 克隆仓库
```bash
git clone https://github.com/Joran1101/meeting-mm.git
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

# Whisper设置
cd ../whisper
pip install -r requirements.txt
```

### 运行

#### 使用一键启动脚本（推荐）

我们提供了便捷的启动和停止脚本，可以一键操作：

```bash
# 启动所有服务（前端+后端）
./start.sh

# 停止所有服务
./stop.sh

# 重启服务
./restart.sh
```

启动脚本会自动：
- 检查并终止已运行的服务
- 启动后端服务（默认在端口8080）
- 启动前端服务（默认在端口3001）
- 所有日志保存在logs目录

#### 手动启动（分别启动各服务）

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

## Notion集成设置

要启用Notion集成，请按照以下步骤操作：

1. 访问 [Notion Integrations](https://www.notion.so/my-integrations) 创建一个新的集成
2. 获取API密钥并添加到 `backend/.env` 文件
3. 在Notion中创建一个新数据库，包含以下属性：
   - Name (标题类型)
   - Date (日期类型)
   - Summary (文本类型)
4. 从数据库URL获取数据库ID (格式为: `https://www.notion.so/[用户名]/[数据库ID]?v=...`)
5. 将数据库ID添加到 `backend/.env` 文件
6. 将你创建的集成与数据库共享，授予"可以编辑"权限

如果遇到Notion集成问题，请参考[Notion集成问题排查指南](docs/troubleshooting/notion_integration_help.md)获取详细的故障排除步骤。

## 测试

项目包含多种测试方式，确保功能正常工作。

### 运行所有测试

使用统一的测试启动脚本运行所有测试：

```bash
./run_tests.sh
```

### 运行特定测试

1. Notion API测试
```bash
./test/scripts/test_notion_connection.sh  # 测试API连接
./test/scripts/test_notion_sync.sh  # 测试同步功能
```

2. 集成测试
```bash
./test/scripts/test_integration.sh
```

3. 后端单元测试
```bash
cd backend
go test ./...
```

4. 前端测试
```bash
cd frontend
npm test
```

### API测试

可以使用curl测试API：

```bash
# 测试分析API
curl -X POST http://localhost:8080/api/meetings/analyze \
  -H "Content-Type: application/json" \
  -d @test/json/test_analyze.json

# 测试Notion同步API
curl -X POST http://localhost:8080/api/meetings/sync-notion \
  -H "Content-Type: application/json" \
  -d @test/json/test_meeting.json
```

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
├── whisper/             # Whisper模型和配置
│   └── models/          # 预训练模型存放目录
├── docs/                # 项目文档
│   └── troubleshooting/ # 故障排除指南
├── logs/                # 服务器日志
├── test/                # 测试相关文件
│   ├── scripts/         # 测试脚本
│   └── json/            # 测试数据
├── start.sh             # 一键启动脚本
├── stop.sh              # 一键停止脚本
├── restart.sh           # 重启脚本
└── run_tests.sh         # 测试启动脚本
```

## 开发日志

开发日志存放在 `logs/` 目录下，记录了每日的开发进度和问题修复情况。详细的开发历程可以查看：

- [2025年3月17日开发日志](logs/development_log_20250317.md) - 项目初始化与基础框架搭建
- [2025年3月18日开发日志](logs/development_log_20250318.md) - 前端重构与核心功能实现
- [2025年3月19日开发日志](logs/development_log_20250319.md) - 功能完善与bug修复
- [2025年3月20日开发日志](logs/development_log_20250320.md) - Notion集成优化与测试完善

完整的变更历史记录在 [CHANGELOG.md](CHANGELOG.md) 文件中。

## 贡献

欢迎提交问题和拉取请求！请访问 [GitHub仓库](https://github.com/Joran1101/meeting-mm) 参与项目开发。

## 联系方式

如有问题或建议，请通过以下方式联系我：
- GitHub: [@Joran1101](https://github.com/Joran1101)
- 邮箱: joran1101@163.com

## 许可证

MIT 