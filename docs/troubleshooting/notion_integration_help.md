# Notion API集成故障排除指南

## 问题分析

当前我们无法成功连接到Notion数据库，错误信息为：
```
Could not find database with ID: 1befcd32-1e77-8093-8562-f8555c1c25e6. Make sure the relevant pages and databases are shared with your integration.
```

这个错误表明：
1. 数据库ID可能是正确的，但数据库没有与集成共享
2. 或者数据库ID不正确

## 解决步骤

### 1. 确认Notion集成设置

1. 访问[Notion的集成页面](https://www.notion.so/my-integrations)
2. 检查您是否已经创建了集成，如果没有，请创建一个新的集成
3. 记录集成的名称和API密钥

### 2. 共享数据库与集成

1. 在Notion中打开您要使用的数据库
2. 在右上角点击"Share"（共享）按钮
3. 在搜索框中输入您刚才创建的集成名称
4. 选择集成并给予"Can edit"（可以编辑）权限
5. 点击"Invite"（邀请）按钮完成共享

### 3. 获取正确的数据库ID

1. 在Notion中打开您的数据库
2. 从浏览器地址栏复制数据库ID：
   - URL格式通常为：`https://www.notion.so/workspace/abc123def456`
   - 或者格式为：`https://www.notion.so/abc123def456?v=...`
   - 数据库ID是`abc123def456`这部分，它应该是32个字符的UUID，可能带有或不带有连字符

### 4. 更新环境配置

1. 在项目的`.env`文件中，更新以下内容：
   ```
   NOTION_API_KEY=您的API密钥
   NOTION_DATABASE_ID=您的数据库ID
   ```

2. 请确保您使用的是带有正确格式的数据库ID：
   - 如果您的ID不包含连字符（如`1befcd321e7780938562f8555c1c25e6`），系统会自动转换为标准UUID格式（`1befcd32-1e77-8093-8562-f8555c1c25e6`）

### 5. 重新测试集成

1. 使用`test_notion_connection.sh`脚本测试API连接
2. 验证您是否可以查询可访问的数据库
3. 如果连接成功，您应该能够看到共享的数据库列表

## 疑难解答提示

- 确保集成和数据库位于同一个Notion工作区
- 检查API密钥是否已过期或无效
- 确认您的Notion账户有足够的权限创建集成
- 如果您使用的是免费计划，可能有API使用限制

## 参考文档

- [Notion API 官方文档](https://developers.notion.com/)
- [Notion API 集成指南](https://developers.notion.com/docs/getting-started)
- [Notion 数据库共享帮助](https://www.notion.so/help/category/import-export-and-integrate) 