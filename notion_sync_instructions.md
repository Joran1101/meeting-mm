# Notion同步问题解决方案

目前后端的Notion API集成遇到了一些格式问题，但我们已找到可工作的解决方案。

## 直接使用脚本同步到Notion

我们已创建了一个可直接工作的脚本`simple_handle_test.sh`，它可以成功地将会议数据同步到Notion数据库。

### 使用方法

1. 确保您已登录Notion并创建了数据库
2. 确保您的数据库有以下属性：
   - `Name` (标题类型)
   - `Date` (日期类型)
   - `Summary` (富文本类型)
3. 确保您已将Notion集成与数据库共享
4. 修改`backend/.env`文件，确保以下内容正确：
   ```
   NOTION_API_KEY=您的Notion API密钥
   NOTION_DATABASE_ID=您的数据库ID (格式为: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
   ```
5. 运行脚本同步会议：
   ```bash
   cd /Users/xujiawei/Downloads/meeting\ mm
   ./simple_handle_test.sh
   ```

### 自定义会议数据

如需自定义会议数据，请编辑`simple_handle_test.sh`文件中的以下部分：

```bash
# 准备简单的数据
TITLE="您的会议标题"
DATE="2025-03-23"
SUMMARY="您的会议摘要"
```

### 故障排除

1. 检查API密钥和数据库ID是否正确
2. 确保数据库中有正确命名的列（Name, Date, Summary）
3. 确保您已将Notion集成与数据库共享
4. 检查日期格式是否为"YYYY-MM-DD"（如"2025-03-23"）

## 后续开发建议

在后续开发中，建议重写后端的Notion集成部分，关键点包括：

1. 确保日期格式正确（作为对象而非字符串）
2. 使用与我们成功脚本相同的请求格式
3. 确保JSON字段名称与Notion API要求匹配
4. 添加更健壮的错误处理和日志记录

这样就能确保Notion同步功能在应用中正常工作。 