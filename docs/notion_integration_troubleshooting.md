# Notion集成问题排查指南

## 前提条件

确保您已完成以下步骤：

1. 在Notion开发者网站[创建了一个集成](https://www.notion.so/my-integrations)
2. 获取了Notion API密钥
3. 创建了一个Notion数据库
4. 将集成与数据库连接（在数据库页面中点击"..."→"添加连接"→选择您的集成）

## 数据库结构

为了使会议纪要正确同步到Notion，数据库必须包含以下属性：

- **Name**：标题类型（Title）- 存储会议标题（必须）
- **Date**：日期类型（Date）- 存储会议日期（必须）
- **Summary**：文本类型（Text）- 存储会议摘要（必须）

如果您的任何字段名称不同或类型不匹配，同步将会失败。

### 特别注意

**Date字段必须是Notion的日期属性类型（Date）**，而不是文本类型。您可以在Notion中通过以下步骤修改字段类型：

1. 打开您的数据库
2. 点击列标题的"..."按钮
3. 选择"属性类型"
4. 选择"日期"

### 字段名称匹配

确保字段名称与代码中期望的完全匹配。大小写敏感！如果您的字段名称是"date"而不是"Date"，或者"name"而不是"Name"，同步将会失败。

## 常见错误

### 1. API密钥或数据库ID无效

错误信息：
```
Notion API密钥未设置或无效
Notion数据库ID未设置或无效
```

解决方案：
- 检查`.env`文件中的`NOTION_API_KEY`和`NOTION_DATABASE_ID`设置
- 确保API密钥已正确复制
- 确保数据库ID是从数据库URL中正确提取的

### 2. 日期格式错误

错误信息：
```
body.properties.Date.date should be an object or null
```

解决方案：
- 确认您的数据库中"Date"字段的类型为"日期"
- 如果类型正确，请检查服务器日志以了解发送的日期格式
- 如果日期字段类型不正确，请在Notion中将其更改为"日期"类型
- 确保日期格式符合ISO 8601标准，格式应为：
  ```json
  "date": {
    "start": "2025-03-20",
    "end": null
  }
  ```

### 3. 权限错误

错误信息：
```
Could not find database with ID: xxx
```

解决方案：
- 确保您的集成已与数据库共享
- 在数据库页面中点击"..."→"添加连接"→选择您的集成

### 4. 字段名称不匹配

错误信息：
```
body.properties.xxxxx should be defined
```

解决方案：
- 检查Notion数据库中的字段名称是否与代码中期望的名称完全匹配
- 可能需要修改代码或重命名Notion数据库列以匹配

## 解决方法

如果您仍然遇到问题，可以尝试以下步骤：

1. **重新创建数据库**：创建全新的数据库，确保所有列名和类型正确
2. **重新创建集成**：创建新的集成并重新获取API密钥
3. **使用Postman测试**：直接使用Postman或curl测试Notion API，了解API期望的确切格式

## 调试步骤

如果您仍然遇到问题，请尝试以下调试步骤：

1. 检查服务器日志中的详细错误信息
2. 验证您的Notion API密钥是否有效（可以使用Postman或curl测试）
3. 确认数据库ID是否正确（可以从URL中提取）
4. 检查数据库的所有必需字段是否存在且类型正确
5. 确保您的集成具有适当的权限

## 支持

如果您在排查后仍然遇到问题，请联系我们的支持团队，并提供以下信息：

- 服务器日志
- Notion数据库的截图（显示字段类型）
- 与Notion API的完整错误响应 

# Notion集成故障排除指南

本文档提供了与Notion API集成过程中可能遇到的常见问题及其解决方案。

## 配置问题

### API密钥无效

**症状**：
- 同步到Notion时出现401未授权错误
- 服务器日志显示"Invalid API key"

**解决方案**：
1. 确保在`.env`文件中正确设置了`NOTION_API_KEY`
2. 验证API密钥是否有效（在Notion的"我的集成"页面检查）
3. 确保API密钥格式正确，通常以`secret_`开头

### 数据库ID无效

**症状**：
- 同步到Notion时出现404未找到错误
- 服务器日志显示"Invalid database ID"

**解决方案**：
1. 确保在`.env`文件中正确设置了`NOTION_DATABASE_ID`
2. 验证数据库ID是否正确（在Notion页面URL或共享链接中查找）
3. 确保已将集成添加到数据库的"连接"中

## API请求问题

### 日期格式验证错误

**症状**：
- 同步到Notion时出现422验证错误
- 错误消息包含`body.properties.Date`相关的验证失败信息
- 错误指出日期应该是对象或null，而不是字符串

**解决方案**：
1. 确保日期格式符合Notion API要求的ISO 8601标准
2. 正确的日期格式应为对象结构：
   ```json
   "date": {
     "start": "2022-02-14",
     "end": null
   }
   ```
3. 在代码中使用`map[string]interface{}`而非`map[string]string`来构建日期属性
4. 确保在请求头中设置了最新的`Notion-Version`（如：`2022-06-28`）

### 数据库字段不匹配

**症状**：
- 同步到Notion时出现属性不存在的错误
- 错误消息指出特定属性在数据库中找不到

**解决方案**：
1. 首先查询数据库结构，了解可用字段：
   ```go
   queryURL := fmt.Sprintf("https://api.notion.com/v1/databases/%s", databaseID)
   req, err := http.NewRequest("GET", queryURL, nil)
   // 设置请求头...
   ```
2. 检查字段名是否完全匹配（Notion区分大小写）
3. 确保属性类型正确（日期字段应使用date类型，文本字段应使用title或rich_text等）
4. 考虑添加自动字段名称映射功能，允许代码查找相似但不完全匹配的字段

## 权限问题

### 无权访问页面

**症状**：
- 同步到Notion时出现403禁止访问错误
- 服务器日志显示"Access denied"或类似消息

**解决方案**：
1. 确保集成已被添加到目标数据库的"连接"中
2. 验证集成权限是否包含读写目标数据库的能力
3. 如果数据库位于工作区内，确保集成已被授权访问该工作区

## 调试技巧

### 启用详细日志

在`services/notion.go`中添加详细的日志记录：

```go
// 请求前记录
log.Printf("正在发送Notion请求: %s", reqBody)

// 响应后记录
log.Printf("Notion响应状态码: %d", resp.StatusCode)
respBody, _ := ioutil.ReadAll(resp.Body)
log.Printf("Notion响应内容: %s", string(respBody))
```

### 使用API版本兼容性

始终在请求中指定Notion API版本：

```go
req.Header.Set("Notion-Version", "2022-06-28")
```

如果API有重大更新，可能需要更新此值。

### 验证API结构

使用Notion API参考文档验证您的请求结构：
https://developers.notion.com/reference

特别注意每种属性类型的特定格式要求。 