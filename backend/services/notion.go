package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"meeting-mm/config"
	"meeting-mm/models"
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

// SyncMeeting 同步会议到Notion
func (s *NotionService) SyncMeeting(meeting *models.Meeting) error {
	// 打印接收到的会议数据以进行调试
	meetingBytes, _ := json.Marshal(meeting)
	log.Printf("【调试】同步到Notion的会议数据: %s\n", string(meetingBytes))

	// 检查标题是否为空
	if meeting == nil {
		return fmt.Errorf("会议对象不能为nil")
	}

	// 打印会议标题
	log.Printf("【调试】会议标题: '%s'，长度: %d\n", meeting.Title, len(meeting.Title))

	// 如果Title为空，设置默认标题
	if meeting.Title == "" {
		meeting.Title = "未命名会议"
		log.Printf("【调试】会议标题为空，已设置默认标题: %s\n", meeting.Title)
	}

	// 如果日期为空，使用当前日期
	if meeting.Date.IsZero() {
		meeting.Date = time.Now()
		log.Printf("会议日期为空，已设置为当前日期: %s\n", meeting.Date.Format("2006-01-02"))
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
						"type": "text",
						"text": map[string]string{
							"content": meeting.Title,
						},
					},
				},
			},
			"Date": map[string]interface{}{
				"date": map[string]interface{}{
					"start": meeting.Date.Format("2006-01-02"),
				},
			},
		},
	}

	// 添加摘要
	if meeting.Summary != "" {
		requestBody["properties"].(map[string]interface{})["Summary"] = map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]string{
						"content": meeting.Summary,
					},
				},
			},
		}
	}

	// 添加参与者
	if len(meeting.Participants) > 0 {
		var participants []map[string]string
		for _, participant := range meeting.Participants {
			participants = append(participants, map[string]string{
				"name": participant,
			})
		}
		requestBody["properties"].(map[string]interface{})["Participants"] = map[string]interface{}{
			"multi_select": participants,
		}
	}

	// 添加页面内容块
	var children []map[string]interface{}

	// 添加摘要部分
	if meeting.Summary != "" {
		children = append(children,
			map[string]interface{}{
				"object": "block",
				"type":   "heading_2",
				"heading_2": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]string{
								"content": "会议摘要",
							},
						},
					},
					"color": "default",
				},
			},
			map[string]interface{}{
				"object": "block",
				"type":   "paragraph",
				"paragraph": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]string{
								"content": meeting.Summary,
							},
						},
					},
					"color": "default",
				},
			},
		)
	}

	// 添加待办事项部分
	if len(meeting.TodoItems) > 0 {
		children = append(children, map[string]interface{}{
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
				"color": "default",
			},
		})

		for _, todo := range meeting.TodoItems {
			todoBlock := map[string]interface{}{
				"object": "block",
				"type":   "to_do",
				"to_do": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]string{
								"content": todo.Description,
							},
						},
					},
					"checked": todo.Status == "completed",
					"color":   "default",
				},
			}

			// 如果有负责人，添加到描述中
			if todo.Assignee != "" {
				todoBlock["to_do"].(map[string]interface{})["rich_text"] = append(
					todoBlock["to_do"].(map[string]interface{})["rich_text"].([]map[string]interface{}),
					map[string]interface{}{
						"type": "text",
						"text": map[string]string{
							"content": fmt.Sprintf(" (@%s)", todo.Assignee),
						},
					},
				)
			}

			// 如果有截止日期，添加到描述中
			if !todo.DueDate.IsZero() {
				todoBlock["to_do"].(map[string]interface{})["rich_text"] = append(
					todoBlock["to_do"].(map[string]interface{})["rich_text"].([]map[string]interface{}),
					map[string]interface{}{
						"type": "text",
						"text": map[string]string{
							"content": fmt.Sprintf(" (截止: %s)", todo.DueDate.Format("2006-01-02")),
						},
					},
				)
			}

			children = append(children, todoBlock)
		}
	}

	// 添加决策事项部分
	if len(meeting.Decisions) > 0 {
		children = append(children, map[string]interface{}{
			"object": "block",
			"type":   "heading_2",
			"heading_2": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"type": "text",
						"text": map[string]string{
							"content": "决策事项",
						},
					},
				},
				"color": "default",
			},
		})

		for _, decision := range meeting.Decisions {
			decisionBlock := map[string]interface{}{
				"object": "block",
				"type":   "bulleted_list_item",
				"bulleted_list_item": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]string{
								"content": decision.Description,
							},
						},
					},
					"color": "default",
				},
			}

			// 如果有决策人，添加到描述中
			if decision.MadeBy != "" {
				decisionBlock["bulleted_list_item"].(map[string]interface{})["rich_text"] = append(
					decisionBlock["bulleted_list_item"].(map[string]interface{})["rich_text"].([]map[string]interface{}),
					map[string]interface{}{
						"type": "text",
						"text": map[string]string{
							"content": fmt.Sprintf(" (由 %s 决定)", decision.MadeBy),
						},
					},
				)
			}

			children = append(children, decisionBlock)
		}
	}

	// 添加会议记录部分
	if meeting.Transcript != "" {
		children = append(children,
			map[string]interface{}{
				"object": "block",
				"type":   "heading_2",
				"heading_2": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]string{
								"content": "会议记录",
							},
						},
					},
					"color": "default",
				},
			},
			map[string]interface{}{
				"object": "block",
				"type":   "paragraph",
				"paragraph": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]string{
								"content": meeting.Transcript,
							},
						},
					},
					"color": "default",
				},
			},
		)
	}

	// 添加内容块到请求体
	if len(children) > 0 {
		requestBody["children"] = children
	}

	// 转换为JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 打印请求体的JSON以便调试
	log.Printf("发送到Notion的请求体: %s\n", string(jsonBody))

	// 创建请求
	req, err := http.NewRequest("POST", "https://api.notion.com/v1/pages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API请求失败，状态码: %d，响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 保存Notion页面ID
	meeting.NotionPageID = response.ID
	log.Printf("会议已成功同步到Notion，页面ID: %s\n", response.ID)

	return nil
}

// UpdateMeetingTranscript 更新会议转录内容
func (s *NotionService) UpdateMeetingTranscript(meeting *models.Meeting) error {
	if meeting.NotionPageID == "" || meeting.Transcript == "" {
		return fmt.Errorf("无效的会议ID或转录内容")
	}

	// 生成转录内容
	var content strings.Builder
	content.WriteString("# 会议转录\n\n")
	content.WriteString(meeting.Transcript)

	// 构建请求体
	requestBody := map[string]interface{}{
		"children": []map[string]interface{}{
			{
				"object": "block",
				"type":   "paragraph",
				"paragraph": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]string{
								"content": content.String(),
							},
						},
					},
				},
			},
		},
	}

	// 转换为JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("PATCH", fmt.Sprintf("https://api.notion.com/v1/blocks/%s/children", meeting.NotionPageID), bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("API请求失败，状态码：%d，响应：%s", resp.StatusCode, string(body))
	}

	return nil
}
