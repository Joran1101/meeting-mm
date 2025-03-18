package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"meeting-mm/config"

	"github.com/google/uuid"
)

// DeepSeekService 提供DeepSeek API调用功能
type DeepSeekService struct {
	apiKey  string
	apiBase string
	client  *http.Client
}

// TodoItem 表示待办事项
type TodoItem struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Assignee    string `json:"assignee"`
	DueDate     string `json:"dueDate"`
	Status      string `json:"status"`
}

// Decision 表示决策点
type Decision struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	MadeBy      string `json:"madeBy"`
}

// Meeting 表示会议
type Meeting struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Date         string     `json:"date"`
	Participants []string   `json:"participants"`
	Transcript   string     `json:"transcript"`
	Summary      string     `json:"summary"`
	TodoItems    []TodoItem `json:"todoItems"`
	Decisions    []Decision `json:"decisions"`
	CreatedAt    string     `json:"createdAt"`
	UpdatedAt    string     `json:"updatedAt"`
	NotionPageID string     `json:"notionPageId,omitempty"`
}

// NewDeepSeekService 创建DeepSeekService实例
func NewDeepSeekService(cfg *config.Config) *DeepSeekService {
	return &DeepSeekService{
		apiKey:  cfg.DeepSeekAPIKey,
		apiBase: cfg.DeepSeekBaseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Message 表示对话消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 表示聊天请求
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

// ChatResponse 表示聊天响应
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// AnalyzeTranscript 分析会议记录，提取待办事项和决策点
func (s *DeepSeekService) AnalyzeTranscript(title, transcript string) (summary string, todoItems []TodoItem, decisions []Decision, err error) {
	// 构建提示词
	prompt := fmt.Sprintf(`你是一个专业的会议纪要助手，请分析以下会议记录，提取关键信息：

会议标题：%s

会议记录：
%s

请提供以下信息：
1. 会议摘要（不超过200字）
2. 待办事项列表（包括负责人和截止日期，如果有的话）
3. 决策点列表（包括决策者，如果有的话）

请以JSON格式返回，格式如下：
{
  "summary": "会议摘要",
  "todoItems": [
    {
      "description": "待办事项描述",
      "assignee": "负责人",
      "dueDate": "截止日期（YYYY-MM-DD格式，如果没有则为null）"
    }
  ],
  "decisions": [
    {
      "description": "决策点描述",
      "madeBy": "决策者（如果没有则为null）"
    }
  ]
}

只返回JSON格式的结果，不要有其他文字。`, title, transcript)

	// 构建请求
	messages := []Message{
		{
			Role:    "system",
			Content: "你是一个专业的会议纪要助手，擅长分析会议记录并提取关键信息。",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	chatReq := ChatRequest{
		Model:       "deepseek-chat",
		Messages:    messages,
		Temperature: 0.3,
		MaxTokens:   2000,
	}

	reqBody, err := json.Marshal(chatReq)
	if err != nil {
		return "", nil, nil, err
	}

	// 发送请求
	req, err := http.NewRequest("POST", s.apiBase+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", nil, nil, fmt.Errorf("API请求失败，状态码：%d，响应：%s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", nil, nil, err
	}

	if len(chatResp.Choices) == 0 {
		return "", nil, nil, errors.New("API返回结果为空")
	}

	// 解析JSON响应
	content := chatResp.Choices[0].Message.Content
	content = strings.TrimSpace(content)

	// 如果返回的内容被包裹在```json和```之间，去掉这些标记
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
	}

	content = strings.TrimSpace(content)

	// 解析JSON
	var result struct {
		Summary   string `json:"summary"`
		TodoItems []struct {
			Description string `json:"description"`
			Assignee    string `json:"assignee"`
			DueDate     string `json:"dueDate"`
		} `json:"todoItems"`
		Decisions []struct {
			Description string `json:"description"`
			MadeBy      string `json:"madeBy"`
		} `json:"decisions"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return "", nil, nil, fmt.Errorf("解析API响应失败：%v，内容：%s", err, content)
	}

	// 转换为返回格式
	todoItems = make([]TodoItem, len(result.TodoItems))
	for i, item := range result.TodoItems {
		todoItems[i] = TodoItem{
			ID:          uuid.New().String(),
			Description: item.Description,
			Assignee:    item.Assignee,
			DueDate:     item.DueDate,
			Status:      "pending",
		}
	}

	decisions = make([]Decision, len(result.Decisions))
	for i, decision := range result.Decisions {
		decisions[i] = Decision{
			ID:          uuid.New().String(),
			Description: decision.Description,
			MadeBy:      decision.MadeBy,
		}
	}

	return result.Summary, todoItems, decisions, nil
}

// GenerateMarkdownReport 生成Markdown格式的会议纪要
func (s *DeepSeekService) GenerateMarkdownReport(meeting Meeting) (string, error) {
	// 构建提示词
	prompt := fmt.Sprintf(`请根据以下会议信息，生成一份完整的Markdown格式会议纪要：

会议标题：%s
会议日期：%s
参与人员：%s
会议摘要：%s

待办事项：
%s

决策点：
%s

会议记录：
%s

请生成一份专业的Markdown格式会议纪要，包括以下部分：
1. 标题和基本信息
2. 摘要
3. 待办事项（使用复选框格式）
4. 决策点
5. 详细会议记录

只返回Markdown格式的结果，不要有其他文字。`,
		meeting.Title,
		meeting.Date,
		strings.Join(meeting.Participants, "、"),
		meeting.Summary,
		formatTodoItems(meeting.TodoItems),
		formatDecisions(meeting.Decisions),
		meeting.Transcript)

	// 构建请求
	messages := []Message{
		{
			Role:    "system",
			Content: "你是一个专业的会议纪要助手，擅长生成格式清晰的Markdown会议纪要。",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	chatReq := ChatRequest{
		Model:       "deepseek-chat",
		Messages:    messages,
		Temperature: 0.3,
		MaxTokens:   4000,
	}

	reqBody, err := json.Marshal(chatReq)
	if err != nil {
		return "", err
	}

	// 发送请求
	req, err := http.NewRequest("POST", s.apiBase+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API请求失败，状态码：%d，响应：%s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", errors.New("API返回结果为空")
	}

	// 获取Markdown内容
	markdown := chatResp.Choices[0].Message.Content
	markdown = strings.TrimSpace(markdown)

	// 如果返回的内容被包裹在```markdown和```之间，去掉这些标记
	if strings.HasPrefix(markdown, "```markdown") {
		markdown = strings.TrimPrefix(markdown, "```markdown")
		markdown = strings.TrimSuffix(markdown, "```")
	} else if strings.HasPrefix(markdown, "```") {
		markdown = strings.TrimPrefix(markdown, "```")
		markdown = strings.TrimSuffix(markdown, "```")
	}

	return strings.TrimSpace(markdown), nil
}

// 格式化待办事项列表
func formatTodoItems(items []TodoItem) string {
	var result strings.Builder
	for _, item := range items {
		if item.DueDate != "" && item.Assignee != "" {
			result.WriteString(fmt.Sprintf("- %s（负责人：%s，截止日期：%s）\n", item.Description, item.Assignee, item.DueDate))
		} else if item.Assignee != "" {
			result.WriteString(fmt.Sprintf("- %s（负责人：%s）\n", item.Description, item.Assignee))
		} else if item.DueDate != "" {
			result.WriteString(fmt.Sprintf("- %s（截止日期：%s）\n", item.Description, item.DueDate))
		} else {
			result.WriteString(fmt.Sprintf("- %s\n", item.Description))
		}
	}
	return result.String()
}

// 格式化决策点列表
func formatDecisions(decisions []Decision) string {
	var result strings.Builder
	for _, decision := range decisions {
		if decision.MadeBy != "" {
			result.WriteString(fmt.Sprintf("- %s（决策者：%s）\n", decision.Description, decision.MadeBy))
		} else {
			result.WriteString(fmt.Sprintf("- %s\n", decision.Description))
		}
	}
	return result.String()
}
