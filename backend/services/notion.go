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

	// 转换数据库ID为UUID格式(如果需要)
	if len(databaseID) == 32 && !strings.Contains(databaseID, "-") {
		formattedID := fmt.Sprintf("%s-%s-%s-%s-%s",
			databaseID[0:8],
			databaseID[8:12],
			databaseID[12:16],
			databaseID[16:20],
			databaseID[20:32])
		log.Printf("数据库ID格式转换: %s -> %s\n", databaseID, formattedID)
		databaseID = formattedID
	}

	// 构建待办事项和决策文本
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

	// 使用硬编码的字段名
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
	client := &http.Client{}
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
		// 输出更详细的错误信息
		log.Printf("API请求失败，状态码：%d，响应：%s\n", resp.StatusCode, string(body))

		// 尝试解析错误消息并提供更多上下文
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				log.Printf("错误消息: %s\n", errMsg)

				// 检查是否是日期格式问题
				if strings.Contains(errMsg, "date should be an object") {
					log.Printf("这可能是一个日期格式问题。应确保日期是一个对象，而不是字符串。")

					// 打印使用的日期字段值
					dateVal := ""
					if props, ok := requestBody["properties"].(map[string]interface{}); ok {
						if dateField, ok := props["Date"].(map[string]interface{}); ok {
							if date, ok := dateField["date"].(map[string]interface{}); ok {
								if start, ok := date["start"].(string); ok {
									dateVal = start
								}
							}
						}
					}
					log.Printf("使用的日期值: %s", dateVal)
				}
			}
		}

		return fmt.Errorf("API请求失败，状态码：%d，响应：%s", resp.StatusCode, string(body))
	}

	log.Println("成功同步到Notion")
	return nil
}

// SyncToNotion 是NotionService的方法，用于同步会议数据到Notion
func (s *NotionService) SyncToNotion(meeting Meeting, markdownReport string) (string, error) {
	// 解析日期
	parsedTime := time.Now()
	var err error
	if meeting.Date != "" {
		// 尝试多种日期格式
		formats := []string{
			"2006-01-02",
			"2006/01/02",
			"2006年01月02日",
			time.RFC3339,
		}

		parsed := false
		for _, format := range formats {
			parsedTime, err = time.Parse(format, meeting.Date)
			if err == nil {
				parsed = true
				fmt.Printf("成功使用格式 '%s' 解析日期\n", format)
				break
			}
		}

		if !parsed {
			fmt.Printf("所有日期格式解析失败，使用当前日期\n")
			parsedTime = time.Now()
		}
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

	// 构建待办事项和决策文本
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
	full_content := meeting.Summary
	if todos_text != "" {
		full_content += "\n\n" + todos_text
	}
	if decisions_text != "" {
		full_content += "\n\n" + decisions_text
	}

	// 构建请求体
	requestBody := map[string]interface{}{
		"parent": map[string]interface{}{
			"database_id": s.databaseID,
		},
		"properties": map[string]interface{}{
			"Name": map[string]interface{}{
				"title": []map[string]interface{}{
					{
						"text": map[string]interface{}{
							"content": meeting.Title,
						},
					},
				},
			},
			"Date": map[string]interface{}{
				"date": map[string]interface{}{
					"start": parsedTime.Format("2006-01-02"),
				},
			},
			"Summary": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"text": map[string]interface{}{
							"content": full_content,
						},
					},
				},
			},
		},
	}

	// 将请求体转换为JSON字符串
	requestBodyStr, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("构建请求体失败: %v", err)
	}

	fmt.Printf("发送到Notion的请求体: %s\n", string(requestBodyStr))

	// 创建请求
	req, err := http.NewRequest("POST", "https://api.notion.com/v1/pages", bytes.NewBuffer(requestBodyStr))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查响应状态
	if resp.StatusCode != 200 {
		fmt.Printf("API请求失败，状态码：%d，响应：%s\n", resp.StatusCode, string(body))

		// 尝试解析错误消息
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errMsg, ok := errorResponse["message"].(string); ok {
				fmt.Printf("错误消息: %s\n", errMsg)
				return "", fmt.Errorf("API请求失败: %s", errMsg)
			}
		}

		return "", fmt.Errorf("API请求失败，状态码：%d，响应：%s", resp.StatusCode, string(body))
	}

	fmt.Printf("成功同步到Notion\n")

	// 从响应获取页面ID
	var pageResponse struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &pageResponse); err != nil {
		return "notion-page-created", nil
	}

	return pageResponse.ID, nil
}

// ApiKey 返回Notion API密钥
func (s *NotionService) ApiKey() string {
	return s.apiKey
}

// DatabaseID 返回Notion数据库ID
func (s *NotionService) DatabaseID() string {
	return s.databaseID
}
