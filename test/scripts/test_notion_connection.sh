#!/bin/bash

# 从环境文件加载Notion API密钥
source ./backend/.env

# 确保变量被正确加载
echo "测试Notion API连通性..."
echo "使用的API密钥: ${NOTION_API_KEY:0:5}...${NOTION_API_KEY: -5}"

# 获取用户信息
echo -e "\n查询用户信息..."
USER_RESPONSE=$(curl -s -X GET \
  -H "Authorization: Bearer $NOTION_API_KEY" \
  -H "Content-Type: application/json" \
  -H "Notion-Version: 2022-06-28" \
  "https://api.notion.com/v1/users")

# 检查响应是否包含错误
if echo "$USER_RESPONSE" | grep -q "\"status\""; then
  echo "API连接失败:"
  echo "$USER_RESPONSE" | grep -o "\"message\":\"[^\"]*\"" | cut -d'"' -f4
  exit 1
else
  echo "API连接成功!"
  echo "返回的用户数据:"
  echo "$USER_RESPONSE" | grep -o "\"name\":\"[^\"]*\"" | cut -d'"' -f4
fi

# 搜索所有可访问的数据库
echo -e "\n搜索可访问的数据库..."
SEARCH_RESPONSE=$(curl -s -X POST \
  -H "Authorization: Bearer $NOTION_API_KEY" \
  -H "Content-Type: application/json" \
  -H "Notion-Version: 2022-06-28" \
  -d '{"filter": {"value": "database", "property": "object"}}' \
  "https://api.notion.com/v1/search")

# 提取所有可访问的数据库ID
echo "找到的数据库:"
echo "$SEARCH_RESPONSE" | grep -o "\"id\":\"[^\"]*\"" | cut -d'"' -f4

# 提取所有数据库标题
echo -e "\n可访问的数据库及其ID:"
DB_IDS=$(echo "$SEARCH_RESPONSE" | grep -o "\"id\":\"[^\"]*\"" | cut -d'"' -f4)
for DB_ID in $DB_IDS; do
  TITLE=$(echo "$SEARCH_RESPONSE" | grep -A50 "$DB_ID" | grep -m1 -o "\"title\":\[.*\]" | awk -F'"content":"' '{print $2}' | cut -d'"' -f1)
  echo "标题: $TITLE, ID: $DB_ID"
done 