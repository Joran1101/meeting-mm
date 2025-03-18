package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"meeting-mm/config"
)

// NotionService 提供Notion API调用功能
type NotionService struct {
	apiKey     string
	databaseID string
	client     *http.Client
}

// NewNotionService 创建NotionService实例
func NewNotionService(cfg *config.Config) *NotionService {
	return &NotionService{
		apiKey:     cfg.NotionAPIKey,
		databaseID: cfg.NotionDatabaseID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NotionPage 表示Notion页面
type NotionPage struct {
	ID string `json:"id"`
}

// 将字符串日期转换为Notion适用的日期对象
func formatDateForNotion(dateStr string) map[string]interface{} {
	return map[string]interface{}{
		"start": dateStr,
	}
}

// SyncToNotion 将会议纪要同步到Notion
func (s *NotionService) SyncToNotion(meeting Meeting, markdownReport string) (string, error) {
	// 打印接收到的会议对象
	meetingBytes, _ := json.Marshal(meeting)
	fmt.Printf("接收到的会议对象: %s\n", string(meetingBytes))

	// 检查API密钥和数据库ID
	if s.apiKey == "" || s.apiKey == "your_notion_api_key_here" {
		return "", fmt.Errorf("Notion API密钥未设置或无效")
	}
	if s.databaseID == "" || s.databaseID == "your_notion_database_id_here" {
		return "", fmt.Errorf("Notion数据库ID未设置或无效")
	}

	// 构建请求体
	requestBody := map[string]interface{}{
		"parent": map[string]string{
			"database_id": s.databaseID,
		},
		"properties": map[string]interface{}{
			"Name": map[string]interface{}{
				"title": []map[string]interface{}{
					{
						"text": map[string]string{
							"content": meeting.Title,
						},
					},
				},
			},
			"Date": map[string]interface{}{
				"date": formatDateForNotion(meeting.Date),
			},
			"Summary": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"text": map[string]string{
							"content": meeting.Summary,
						},
					},
				},
			},
		},
		"children": []map[string]interface{}{
			{
				"object": "block",
				"type":   "paragraph",
				"paragraph": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]string{
								"content": "以下是会议纪要的详细内容：",
							},
						},
					},
				},
			},
			{
				"object": "block",
				"type":   "toggle",
				"toggle": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]string{
								"content": "会议纪要",
							},
						},
					},
					"children": []map[string]interface{}{
						{
							"object": "block",
							"type":   "paragraph",
							"paragraph": map[string]interface{}{
								"rich_text": []map[string]interface{}{
									{
										"type": "text",
										"text": map[string]string{
											"content": markdownReport,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// 添加待办事项
	if len(meeting.TodoItems) > 0 {
		todoBlock := map[string]interface{}{
			"object": "block",
			"type":   "heading_2",
			"heading_2": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"type": "text",
						"text": map[string]string{
							"content": "待办事项",
						},
					},
				},
			},
		}
		requestBody["children"] = append(requestBody["children"].([]map[string]interface{}), todoBlock)

		for _, item := range meeting.TodoItems {
			todoItemBlock := map[string]interface{}{
				"object": "block",
				"type":   "to_do",
				"to_do": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]string{
								"content": fmt.Sprintf("%s（%s）", item.Description, item.Assignee),
							},
						},
					},
					"checked": item.Status == "completed",
				},
			}
			requestBody["children"] = append(requestBody["children"].([]map[string]interface{}), todoItemBlock)
		}
	}

	// 添加决策点
	if len(meeting.Decisions) > 0 {
		decisionBlock := map[string]interface{}{
			"object": "block",
			"type":   "heading_2",
			"heading_2": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"type": "text",
						"text": map[string]string{
							"content": "决策点",
						},
					},
				},
			},
		}
		requestBody["children"] = append(requestBody["children"].([]map[string]interface{}), decisionBlock)

		for _, decision := range meeting.Decisions {
			decisionText := decision.Description
			if decision.MadeBy != "" {
				decisionText = fmt.Sprintf("%s（决策者：%s）", decision.Description, decision.MadeBy)
			}

			decisionItemBlock := map[string]interface{}{
				"object": "block",
				"type":   "bulleted_list_item",
				"bulleted_list_item": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]string{
								"content": decisionText,
							},
						},
					},
				},
			}
			requestBody["children"] = append(requestBody["children"].([]map[string]interface{}), decisionItemBlock)
		}
	}

	// 序列化请求体
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 打印请求体
	fmt.Printf("发送到Notion的请求体: %s\n", string(jsonData))

	// 创建请求
	req, err := http.NewRequest("POST", "https://api.notion.com/v1/pages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API请求失败，状态码：%d，响应：%s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	var notionPage NotionPage
	if err := json.NewDecoder(resp.Body).Decode(&notionPage); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	return notionPage.ID, nil
}
