# 测试说明

本目录包含会议纪要系统的各种测试脚本和测试数据。

## 目录结构

- `scripts/`: 包含各种测试脚本
  - `integration_test.sh`: 集成测试脚本，测试前端和后端的整合
  - `notion_test.sh`: Notion API同步测试脚本
  - `test_notion_connection.sh`: Notion API连接测试脚本
  
- `json/`: 包含用于测试的JSON数据文件
  - `test_analyze.json`: 用于测试会议分析API的数据
  - `test_meeting.json`: 包含完整会议数据的测试文件

## 如何使用测试脚本

### 集成测试

运行集成测试以验证整个系统的功能：

```bash
cd <项目根目录>
./test/scripts/integration_test.sh
```

### Notion API测试

测试Notion API连接和同步功能：

```bash
cd <项目根目录>
./test/scripts/test_notion_connection.sh  # 测试API连接
./test/scripts/notion_test.sh  # 测试同步功能
```

### 使用JSON文件进行测试

通过curl或其他HTTP客户端使用JSON文件测试API：

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

## 注意事项

1. 确保在测试前配置了必要的环境变量（在`backend/.env`文件中）
2. 测试前确保后端服务已启动
3. 如果测试失败，检查日志文件`backend/backend_log.txt`获取详细信息 