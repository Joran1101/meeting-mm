package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"meeting-mm/config"

	"github.com/joho/godotenv"
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
	// 尝试解析日期字符串
	parsedTime, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		// 尝试其他常见格式
		parsedTime, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			parsedTime, err = time.Parse("2006/01/02", dateStr)
			if err != nil {
				parsedTime, err = time.Parse("2006-01-02T15:04:05Z", dateStr)
				if err != nil {
					// 如果无法解析，返回当前日期
					fmt.Printf("无法解析日期 '%s': %v, 使用当前日期\n", dateStr, err)
					parsedTime = time.Now()
				}
			}
		}
	}

	// 转换为Notion需要的ISO 8601格式 (YYYY-MM-DD)
	formattedDate := parsedTime.Format("2006-01-02")

	return map[string]interface{}{
		"start": formattedDate,
	}
}

// SyncToNotion 将会议摘要同步到Notion
func SyncToNotion(meetingTitle string, meetingDate time.Time, meetingSummary string, todos []string, decisions []string) error {
	// 加载环境变量
	if err := godotenv.Load(".env"); err != nil {
		log.Println("错误: 无法加载.env文件:", err)
		// 尝试从父目录加载
		if err := godotenv.Load("../.env"); err != nil {
			return errors.New("无法加载环境变量")
		}
	}

	apiKey := os.Getenv("NOTION_API_KEY")
	databaseID := os.Getenv("NOTION_DATABASE_ID")

	if apiKey == "" || strings.Contains(apiKey, "您的Notion_API密钥") {
		return errors.New("Notion API密钥未设置或无效")
	}

	if databaseID == "" || strings.Contains(databaseID, "您的Notion数据库ID") {
		return errors.New("Notion数据库ID未设置或无效")
	}

	// 首先查询数据库结构，确保字段存在且类型正确
	log.Println("查询Notion数据库结构...")
	dbURL := fmt.Sprintf("https://api.notion.com/v1/databases/%s", databaseID)
	dbReq, err := http.NewRequest("GET", dbURL, nil)
	if err != nil {
		log.Println("创建数据库查询请求失败:", err)
		return err
	}

	// 设置请求头
	dbReq.Header.Set("Authorization", "Bearer "+apiKey)
	dbReq.Header.Set("Content-Type", "application/json")
	dbReq.Header.Set("Notion-Version", "2022-06-28")

	// 发送请求
	client := &http.Client{}
	dbResp, err := client.Do(dbReq)
	if err != nil {
		log.Println("发送数据库查询请求失败:", err)
		return err
	}
	defer dbResp.Body.Close()

	// 读取响应
	dbBody, err := ioutil.ReadAll(dbResp.Body)
	if err != nil {
		log.Println("读取数据库响应失败:", err)
		return err
	}

	// 检查响应状态
	if dbResp.StatusCode != 200 {
		log.Printf("数据库查询失败，状态码: %d, 响应: %s\n", dbResp.StatusCode, string(dbBody))
		return fmt.Errorf("数据库查询失败，状态码: %d", dbResp.StatusCode)
	}

	// 解析响应
	var dbData map[string]interface{}
	if err := json.Unmarshal(dbBody, &dbData); err != nil {
		log.Println("解析数据库响应失败:", err)
		return err
	}

	// 检查数据库属性
	properties, ok := dbData["properties"].(map[string]interface{})
	if !ok {
		log.Println("无法获取数据库属性")
		return errors.New("无法获取数据库属性")
	}

	// 输出数据库结构以便调试
	log.Println("数据库属性:")
	for name, prop := range properties {
		propType := ""
		if propMap, ok := prop.(map[string]interface{}); ok {
			if typeVal, ok := propMap["type"].(string); ok {
				propType = typeVal
			}
		}
		log.Printf("属性名: %s, 类型: %s\n", name, propType)
	}

	// 检查必要的字段是否存在
	requiredFields := map[string]string{
		"Name":    "title",
		"Date":    "date",
		"Summary": "rich_text",
	}

	for field, expectedType := range requiredFields {
		prop, ok := properties[field]
		if !ok {
			log.Printf("警告: 数据库中缺少必要字段 %s\n", field)
			// 如果字段名不匹配，尝试查找类似的字段
			for name, _ := range properties {
				if strings.EqualFold(name, field) {
					log.Printf("找到类似字段: %s，可能需要重命名为 %s\n", name, field)
				}
			}
			continue
		}

		propMap, ok := prop.(map[string]interface{})
		if !ok {
			log.Printf("警告: 属性 %s 格式不正确\n", field)
			continue
		}

		actualType, ok := propMap["type"].(string)
		if !ok || actualType != expectedType {
			log.Printf("警告: 属性 %s 的类型应为 %s，实际为 %s\n", field, expectedType, actualType)
		}
	}

	// 构建请求体
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

	full_content := meetingSummary
	if todos_text != "" {
		full_content += "\n\n" + todos_text
	}
	if decisions_text != "" {
		full_content += "\n\n" + decisions_text
	}

	// 正确格式化日期（符合ISO 8601）
	formattedDate := meetingDate.Format("2006-01-02")
	log.Printf("使用的日期格式: %s\n", formattedDate)

	// 构建请求体
	requestBody := map[string]interface{}{
		"parent": map[string]string{
			"database_id": databaseID,
		},
		"properties": map[string]interface{}{
			"Name": map[string]interface{}{
				"title": []map[string]interface{}{
					{
						"text": map[string]string{
							"content": meetingTitle,
						},
					},
				},
			},
			"Date": map[string]interface{}{
				"date": map[string]interface{}{
					"start": formattedDate,
					"end":   nil,
				},
			},
			"Summary": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"text": map[string]string{
							"content": full_content,
						},
					},
				},
			},
		},
	}

	// 转换为JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	log.Println("发送到Notion的请求体:", string(jsonBody))

	// 创建请求
	req, err := http.NewRequest("POST", "https://api.notion.com/v1/pages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 检查响应状态
	if resp.StatusCode != 200 {
		log.Printf("Notion同步失败，状态码: %d, 响应: %s\n", resp.StatusCode, string(body))
		return fmt.Errorf("Notion同步失败，状态码: %d", resp.StatusCode)
	}

	log.Println("成功同步到Notion")
	return nil
}

// SyncToNotion 是NotionService的方法，用于同步会议数据到Notion
func (s *NotionService) SyncToNotion(meeting Meeting, markdownReport string) (string, error) {
	// 解析日期
	parsedTime, err := time.Parse("2006-01-02", meeting.Date)
	if err != nil {
		fmt.Printf("解析日期失败: %v，使用当前日期\n", err)
		parsedTime = time.Now()
	}

	// 提取待办事项和决策点
	var todos []string
	var decisions []string

	for _, todo := range meeting.TodoItems {
		todos = append(todos, todo.Description)
	}

	for _, decision := range meeting.Decisions {
		decisions = append(decisions, decision.Description)
	}

	// 调用全局SyncToNotion函数
	err = SyncToNotion(
		meeting.Title,
		parsedTime,
		meeting.Summary,
		todos,
		decisions,
	)

	if err != nil {
		return "", err
	}

	// 由于全局函数不返回PageID，这里只返回一个空字符串作为占位符
	return "notion-page-synced", nil
}
