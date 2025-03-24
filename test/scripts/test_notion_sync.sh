#!/bin/bash

# 加载环境变量
source ./backend/.env

echo "测试 Notion API 集成..."
echo "API Key: ${NOTION_API_KEY:0:10}..."
echo "Database ID: ${NOTION_DATABASE_ID:0:10}..."

# 构建测试数据
TEST_DATA='{
  "id": "test-001",
  "title": "测试会议 - Notion集成",
  "date": "2024-03-23",
  "participants": ["测试用户1", "测试用户2"],
  "transcript": "这是一个测试会议的记录",
  "summary": "测试会议总结",
  "todo_items": [
    {
      "description": "测试待办事项1"
    },
    {
      "description": "测试待办事项2"
    }
  ],
  "decisions": [
    {
      "description": "测试决策1"
    }
  ],
  "created_at": "2024-03-23T10:00:00Z",
  "updated_at": "2024-03-23T10:00:00Z"
}'

# 发送请求到后端
echo "发送测试数据到后端..."
curl -X POST http://localhost:8080/api/meetings/sync-notion \
  -H "Content-Type: application/json" \
  -d "$TEST_DATA"

echo -e "\n测试完成。请检查 Notion 数据库是否已创建新页面。" 