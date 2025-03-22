package services

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// SyncViaScript 通过直接运行shell脚本与Notion API交互
// 这是一个可靠的方式来同步会议数据到Notion，避免了日期格式问题
func (s *NotionService) SyncViaScript(meeting Meeting) (string, error) {
	// 准备内容
	title := meeting.Title
	date := meeting.Date
	summary := meeting.Summary

	// 构建待办事项和决策文本
	var todos []string
	var decisions []string
	for _, todo := range meeting.TodoItems {
		todos = append(todos, todo.Description)
	}
	for _, decision := range meeting.Decisions {
		decisions = append(decisions, decision.Description)
	}

	todos_text := ""
	if len(todos) > 0 {
		todos_text = "### 待办事项\n\n"
		for _, todo := range todos {
			todos_text += fmt.Sprintf("- [ ] %s\n", todo)
		}
	}

	decisions_text := ""
	if len(decisions) > 0 {
		decisions_text += "### 决策\n\n"
		for _, decision := range decisions {
			decisions_text += fmt.Sprintf("- %s\n", decision)
		}
	}

	// 合并内容
	full_content := summary
	if todos_text != "" {
		full_content += "\n\n" + todos_text
	}
	if decisions_text != "" {
		full_content += "\n\n" + decisions_text
	}

	// 创建临时脚本文件
	scriptFile, err := os.CreateTemp("", "notion_sync_*.sh")
	if err != nil {
		return "", fmt.Errorf("创建临时脚本失败: %v", err)
	}
	defer os.Remove(scriptFile.Name()) // 确保删除临时脚本文件

	// 准备脚本内容
	scriptContent := fmt.Sprintf(`#!/bin/bash
# 使用提供的参数
NOTION_API_KEY="%s"
NOTION_DATABASE_ID="%s"
TITLE="%s"
DATE="%s"
SUMMARY="%s"

# 输出参数
echo "执行脚本使用参数:"
echo "API密钥: ${NOTION_API_KEY:0:5}...${NOTION_API_KEY: -4}"
echo "数据库ID: $NOTION_DATABASE_ID"
echo "标题: $TITLE"
echo "日期: $DATE" 

# 创建请求体 - 确保格式与成功的案例完全一致
REQUEST_BODY='{
  "parent": {
    "database_id": "'$NOTION_DATABASE_ID'"
  },
  "properties": {
    "Name": {
      "title": [
        {
          "text": {
            "content": "'$TITLE'"
          }
        }
      ]
    },
    "Date": {
      "date": {
        "start": "'$DATE'"
      }
    },
    "Summary": {
      "rich_text": [
        {
          "text": {
            "content": "'$SUMMARY'"
          }
        }
      ]
    }
  }
}'

# 发送请求
RESPONSE=$(curl -s -X POST \
  -H "Authorization: Bearer $NOTION_API_KEY" \
  -H "Content-Type: application/json" \
  -H "Notion-Version: 2022-06-28" \
  -d "$REQUEST_BODY" \
  "https://api.notion.com/v1/pages")

# 检查成功并提取ID
if echo "$RESPONSE" | grep -q "\"object\":\"page\""; then
  PAGE_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  echo "SUCCESS:$PAGE_ID"
  exit 0
else
  echo "ERROR:$RESPONSE"
  exit 1
fi
`, s.apiKey, s.databaseID, escapeShellString(title), escapeShellString(date), escapeShellString(full_content))

	// 写入脚本内容
	if _, err := scriptFile.Write([]byte(scriptContent)); err != nil {
		return "", fmt.Errorf("写入脚本文件失败: %v", err)
	}
	scriptFile.Close()

	// 设置脚本为可执行
	if err := os.Chmod(scriptFile.Name(), 0755); err != nil {
		return "", fmt.Errorf("设置脚本权限失败: %v", err)
	}

	// 执行脚本
	fmt.Printf("执行Notion同步脚本...\n")
	output, err := exec.Command("/bin/bash", scriptFile.Name()).CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return "", fmt.Errorf("执行脚本失败: %v, 输出: %s", err, outputStr)
	}

	// 解析输出以获取页面ID或错误
	if strings.Contains(outputStr, "SUCCESS:") {
		// 提取页面ID
		for _, line := range strings.Split(outputStr, "\n") {
			if strings.Contains(line, "SUCCESS:") {
				parts := strings.SplitN(line, "SUCCESS:", 2)
				if len(parts) > 1 {
					return strings.TrimSpace(parts[1]), nil
				}
				break
			}
		}
		return "notion-page-created", nil
	} else if strings.Contains(outputStr, "ERROR:") {
		// 提取错误信息
		for _, line := range strings.Split(outputStr, "\n") {
			if strings.Contains(line, "ERROR:") {
				parts := strings.SplitN(line, "ERROR:", 2)
				if len(parts) > 1 {
					return "", fmt.Errorf("API请求失败: %s", parts[1])
				}
				break
			}
		}
		return "", fmt.Errorf("未知API错误")
	}

	return "", fmt.Errorf("无法从脚本输出中获取结果")
}

// escapeShellString 转义shell字符串中的特殊字符
func escapeShellString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, `$`, `\$`)
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	return s
}
