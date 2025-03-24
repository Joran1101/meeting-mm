# Notion集成问题排查指南

本文档提供了会议纪要自动生成器与Notion集成时可能遇到的常见问题及解决方案。

## 常见问题

### 1. 同步失败，标题或摘要字段未显示

**问题表现**：
- 会议数据同步到Notion后，标题或摘要字段为空
- 后端日志显示 "API请求失败" 错误

**解决方案**：
- 确保在Notion数据库中存在正确的字段：
  - `Name` 字段 (标题类型)
  - `Summary` 字段 (文本类型)
  - `Date` 字段 (日期类型)
- 检查 `.env` 文件中的 API 密钥和数据库 ID 是否正确
- 检查集成与数据库是否正确共享（需要"可以编辑"权限）

### 2. Notion API 认证错误

**问题表现**：
- 后端日志显示 "401 Unauthorized" 或 "API请求失败，状态码: 401"
- 无法创建或更新Notion页面

**解决方案**：
- 确认 API 密钥是否正确复制到 `.env` 文件
- 检查 API 密钥是否有效（可以在Notion集成页面重新获取）
- 确保 API 密钥没有过期或被撤销
- 重启后端服务以加载最新的环境变量

### 3. 日期格式错误

**问题表现**：
- 同步过程中出现 "日期格式无效" 错误
- Notion中显示的日期不正确

**解决方案**：
- 确保日期格式符合ISO 8601标准 (YYYY-MM-DD)
- 检查前端是否正确处理日期格式
- 如果日期为空，系统会自动使用当前日期

### 4. 数据库ID错误

**问题表现**：
- 后端日志显示 "404 Not Found" 或 "API请求失败，状态码: 404"
- 无法找到目标数据库

**解决方案**：
- 确认数据库ID是否正确（来自数据库URL）
- 格式应为：`https://www.notion.so/[用户名]/[数据库ID]?v=...`
- 数据库ID是32位字符串，中间有短横线
- 确保集成与数据库正确共享

### 5. Participants字段错误

**问题表现**：
- 参与者信息未正确显示
- 后端日志显示与属性相关的错误

**解决方案**：
- 确保Notion数据库中存在`Participants`字段，类型为`multi-select`
- 如果没有此字段，可以手动添加，或修改代码适应现有字段
- 重新同步会议数据

## 调试技巧

### 启用详细日志

修改 `.env` 文件，启用详细日志记录：
```
DEBUG=true
LOG_LEVEL=debug
```

### 检查日志文件

同步问题的详细错误通常记录在日志文件中：
```
tail -f logs/backend_log.txt
```

### 使用测试脚本

使用提供的测试脚本进行API测试：
```
./test/scripts/test_notion_connection.sh  # 测试API连接
./test/scripts/test_notion_sync.sh  # 测试同步功能
```

### 手动验证API请求

使用curl测试Notion API：
```
curl -X POST https://api.notion.com/v1/pages \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Notion-Version: 2022-06-28" \
  -H "Content-Type: application/json" \
  -d '{
    "parent": { "database_id": "YOUR_DATABASE_ID" },
    "properties": {
      "Name": {
        "title": [{ "type": "text", "text": { "content": "测试标题" } }]
      }
    }
  }'
```

## 最近修复

我们最近修复了以下Notion集成问题：

1. **标题和摘要字段同步问题** - 修复了请求格式，确保字段正确同步
2. **路由配置错误** - 修复了API路由映射
3. **日期格式处理** - 改进了日期处理逻辑，支持多种格式
4. **数据兼容性** - 增强了处理函数以支持新旧两种请求格式

如果您遇到其他问题，请[提交问题报告](https://github.com/Joran1101/meeting-mm/issues)或联系开发团队获取帮助。 