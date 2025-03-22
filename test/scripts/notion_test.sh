#!/bin/bash

# 从环境文件加载API密钥和数据库ID
source ./backend/.env

# 显示运行信息
echo "最简单的Notion API测试脚本"
echo "API密钥: ${NOTION_API_KEY:0:5}...${NOTION_API_KEY: -4}"
echo "数据库ID: ${NOTION_DATABASE_ID}"

# 准备简单的数据
TITLE="直接脚本最终测试"
DATE="2025-03-23"
SUMMARY="这是一个解决日期格式问题的最终测试"
echo "标题: $TITLE"
echo "日期: $DATE"
echo "摘要: $SUMMARY"

# 创建请求体 - 确保格式与成功的案例完全一致
# 注意：根据最小化变化原则，我们保持JSON格式完全一致，包括缩进
REQUEST_BODY='{
  "parent": {
    "database_id": "'"$NOTION_DATABASE_ID"'"
  },
  "properties": {
    "Name": {
      "title": [
        {
          "text": {
            "content": "'"$TITLE"'"
          }
        }
      ]
    },
    "Date": {
      "date": {
        "start": "'"$DATE"'"
      }
    },
    "Summary": {
      "rich_text": [
        {
          "text": {
            "content": "'"$SUMMARY"'"
          }
        }
      ]
    }
  }
}'

# 显示请求体
echo -e "\n请求体:"
echo "$REQUEST_BODY"

# 发送请求
echo -e "\n发送请求到Notion API..."
RESPONSE=$(curl -s -X POST \
  -H "Authorization: Bearer $NOTION_API_KEY" \
  -H "Content-Type: application/json" \
  -H "Notion-Version: 2022-06-28" \
  -d "$REQUEST_BODY" \
  "https://api.notion.com/v1/pages")

# 显示响应
if echo "$RESPONSE" | grep -q "id"; then
  echo "成功创建页面!"
  PAGE_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  echo "页面ID: $PAGE_ID"
  echo -e "\n响应开头部分:"
  echo "$RESPONSE" | head -n 10
else
  echo "创建页面失败:"
  echo "$RESPONSE"
fi 