# 会议纪要自动生成器

一个使用DeepSeek和Whisper进行会议纪要自动生成的工具，可以从会议录音中提取行动项和决策点。

## 功能特点

- 🎙️ 会议录音转文字（使用Whisper）
- 🤖 使用DeepSeek提取行动项和决策点
- 📝 生成结构化的会议纪要
- 📊 支持与Notion集成
- 📱 响应式Web界面，支持多种设备

## 技术栈

- **后端**: Go + Fiber
- **前端**: React + TypeScript + Material-UI
- **AI模型**: DeepSeek Chat + Whisper
- **数据存储**: Notion API (可选)

## 开发环境要求

- Go 1.18+
- Node.js 16+
- npm 8+
- Python 3.8+ (用于Whisper)
- DeepSeek API密钥
- Notion API密钥 (可选)
- Whisper模型 (本地运行)

## 项目当前状态

- ✅ 音频上传和转录功能
- ✅ 会议内容分析（摘要、待办事项、决策点）
- ✅ Markdown格式报告生成
- ✅ Notion集成基础功能
- ⏳ 音频文件持久化存储
- ⏳ 历史会议记录管理
- ⏳ UI/UX优化

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

1. 启动后端服务
```bash
cd backend
./meeting-mm-server
# 或者从源码运行
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

## 音频文件管理

当前版本中，音频文件处理流程如下：

1. 上传的音频文件被临时存储在系统临时目录 (`meeting-mm-whisper/`)
2. 文件使用随机UUID命名，扩展名为 `.mp3`
3. 转录完成后文件会被自动删除

音频文件管理将在后续版本中优化，添加持久化存储选项。

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
├── whisper/             # Whisper模型和配置
│   └── models/          # 预训练模型存放目录
└── logs/                # 服务器日志和开发记录
```

## 开发日志

开发日志存放在 `logs/` 目录下，记录了每日的开发进度和问题修复情况。

## 贡献

欢迎提交问题和拉取请求！请访问 [GitHub仓库](https://github.com/Joran1101/meeting-mm) 参与项目开发。

## 联系方式

如有问题或建议，请通过以下方式联系我：
- GitHub: [@Joran1101](https://github.com/Joran1101)
- 邮箱: joran1101@163.com

## 许可证

MIT 