package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"meeting-mm/models"
	"meeting-mm/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler 处理API请求
type Handler struct {
	deepseekService *services.DeepSeekService
	notionService   *services.NotionService
	whisperService  *services.WhisperService
}

// NewHandler 创建Handler实例
func NewHandler(deepseekService *services.DeepSeekService, notionService *services.NotionService, whisperService *services.WhisperService) *Handler {
	return &Handler{
		deepseekService: deepseekService,
		notionService:   notionService,
		whisperService:  whisperService,
	}
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// UploadAudio 上传音频文件
func (h *Handler) UploadAudio(c *fiber.Ctx) error {
	// 获取表单数据
	title := c.FormValue("title")
	if title == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "会议标题不能为空",
		})
	}

	syncToNotion := c.FormValue("syncToNotion") == "true"

	// 获取音频文件
	file, err := c.FormFile("audio")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("获取音频文件失败: %v", err),
		})
	}

	// 打开文件
	fileHandle, err := file.Open()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("打开音频文件失败: %v", err),
		})
	}
	defer fileHandle.Close()

	// 读取文件内容
	audioData, err := io.ReadAll(fileHandle)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("读取音频文件失败: %v", err),
		})
	}

	// 转录音频
	transcript, err := h.whisperService.TranscribeAudio(audioData)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("转录音频失败: %v", err),
		})
	}

	// 分析转录内容
	summary, todoItems, decisions, err := h.deepseekService.AnalyzeTranscript(title, transcript)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("分析转录内容失败: %v", err),
		})
	}

	// 创建会议对象
	meeting := &models.Meeting{
		ID:           uuid.New().String(),
		Title:        title,
		Date:         time.Now(),
		Participants: []string{}, // 这里可以从请求中获取参与者信息
		Transcript:   transcript,
		Summary:      summary,
		TodoItems:    make([]models.TodoItem, len(todoItems)),
		Decisions:    make([]models.Decision, len(decisions)),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 转换待办事项
	for i, todo := range todoItems {
		dueDate, err := time.Parse("2006-01-02", todo.DueDate)
		if err != nil {
			dueDate = time.Time{} // 如果解析失败，使用零值
		}
		meeting.TodoItems[i] = models.TodoItem{
			ID:          todo.ID,
			Description: todo.Description,
			Assignee:    todo.Assignee,
			DueDate:     dueDate,
			Status:      todo.Status,
		}
	}

	// 转换决策点
	for i, decision := range decisions {
		meeting.Decisions[i] = models.Decision{
			ID:          decision.ID,
			Description: decision.Description,
			MadeBy:      decision.MadeBy,
		}
	}

	// 生成Markdown报告
	markdownReport, err := h.deepseekService.GenerateMarkdownReport(services.Meeting{
		ID:           meeting.ID,
		Title:        meeting.Title,
		Date:         meeting.Date.Format("2006-01-02"),
		Participants: meeting.Participants,
		Transcript:   meeting.Transcript,
		Summary:      meeting.Summary,
		TodoItems:    todoItems,
		Decisions:    decisions,
		CreatedAt:    meeting.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    meeting.UpdatedAt.Format(time.RFC3339),
	})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("生成Markdown报告失败: %v", err),
		})
	}

	// 同步到Notion（如果需要）
	if syncToNotion {
		err := h.notionService.SyncMeeting(meeting)
		if err != nil {
			// 不中断流程，只记录错误
			fmt.Printf("同步到Notion失败: %v\n", err)
		}
	}

	// 返回结果
	return c.JSON(fiber.Map{
		"meeting":        meeting,
		"markdownReport": markdownReport,
	})
}

// StreamAudio 流式处理音频
func (h *Handler) StreamAudio(c *fiber.Ctx) error {
	// 获取采样率
	sampleRateStr := c.Query("sampleRate", "16000")
	_, err := strconv.Atoi(sampleRateStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("无效的采样率: %v", err),
		})
	}

	// 读取请求体
	audioStream := c.Request().BodyStream()
	if audioStream == nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "请求体为空",
		})
	}

	// 流式转录
	resultChan, err := h.whisperService.StreamTranscribe(audioStream)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("流式转录失败: %v", err),
		})
	}

	// 从通道读取结果
	transcript := ""
	for result := range resultChan {
		if strings.HasPrefix(result, "错误:") {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": result,
			})
		}
		transcript = result
	}

	// 返回结果
	return c.JSON(fiber.Map{
		"transcript": transcript,
	})
}

// AnalyzeTranscript 分析会议转录并生成会议纪要
func (h *Handler) AnalyzeTranscript(c *fiber.Ctx) error {
	// 解析请求
	var request struct {
		Title      string `json:"title"`
		Transcript string `json:"transcript"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("解析请求体失败: %v", err),
		})
	}

	title := request.Title
	transcript := request.Transcript

	if title == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "会议标题不能为空",
		})
	}

	if transcript == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "会议转录不能为空",
		})
	}

	// 分析转录内容
	summary, todoItems, decisions, err := h.deepseekService.AnalyzeTranscript(title, transcript)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("分析转录内容失败: %v", err),
		})
	}

	// 创建会议对象
	meeting := &models.Meeting{
		ID:           uuid.New().String(),
		Title:        title,
		Date:         time.Now(),
		Participants: []string{}, // 这里可以从请求中获取参与者信息
		Transcript:   transcript,
		Summary:      summary,
		TodoItems:    make([]models.TodoItem, len(todoItems)),
		Decisions:    make([]models.Decision, len(decisions)),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 转换待办事项
	for i, todo := range todoItems {
		dueDate, err := time.Parse("2006-01-02", todo.DueDate)
		if err != nil {
			dueDate = time.Time{} // 如果解析失败，使用零值
		}
		meeting.TodoItems[i] = models.TodoItem{
			ID:          todo.ID,
			Description: todo.Description,
			Assignee:    todo.Assignee,
			DueDate:     dueDate,
			Status:      todo.Status,
		}
	}

	// 转换决策点
	for i, decision := range decisions {
		meeting.Decisions[i] = models.Decision{
			ID:          decision.ID,
			Description: decision.Description,
			MadeBy:      decision.MadeBy,
		}
	}

	// 生成Markdown报告
	markdownReport, err := h.deepseekService.GenerateMarkdownReport(services.Meeting{
		ID:           meeting.ID,
		Title:        meeting.Title,
		Date:         meeting.Date.Format("2006-01-02"),
		Participants: meeting.Participants,
		Transcript:   meeting.Transcript,
		Summary:      meeting.Summary,
		TodoItems:    todoItems,
		Decisions:    decisions,
		CreatedAt:    meeting.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    meeting.UpdatedAt.Format(time.RFC3339),
	})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("生成Markdown报告失败: %v", err),
		})
	}

	// 返回结果
	return c.JSON(fiber.Map{
		"meeting":        meeting,
		"markdownReport": markdownReport,
	})
}

// SyncToNotionHandler 处理将会议数据同步到Notion
func (h *Handler) SyncToNotionHandler(c *fiber.Ctx) error {
	// 打印原始请求体
	requestBody := string(c.Body())
	fmt.Printf("【调试】原始请求体: %s\n", requestBody)

	var request struct {
		Meeting        models.Meeting `json:"meeting"`
		MarkdownReport string         `json:"markdownReport"`
	}

	if err := c.BodyParser(&request); err != nil {
		// 尝试解析旧格式的请求
		var oldRequest struct {
			ID           string   `json:"id"`
			Title        string   `json:"title"`
			Date         string   `json:"date"`
			Participants []string `json:"participants"`
			Transcript   string   `json:"transcript"`
			Summary      string   `json:"summary"`
			TodoItems    []struct {
				Description string `json:"description"`
				Assignee    string `json:"assignee"`
				DueDate     string `json:"dueDate"`
				Status      string `json:"status"`
			} `json:"todo_items"`
			Decisions []struct {
				Description string `json:"description"`
				MadeBy      string `json:"madeBy"`
			} `json:"decisions"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		}

		if err2 := c.BodyParser(&oldRequest); err2 != nil {
			fmt.Printf("解析请求体失败: %v\n", err)
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("解析请求体失败: %v", err),
			})
		}

		// 将旧格式转换为新格式
		meeting := &models.Meeting{
			ID:           oldRequest.ID,
			Title:        oldRequest.Title,
			Participants: oldRequest.Participants,
			Transcript:   oldRequest.Transcript,
			Summary:      oldRequest.Summary,
		}

		// 解析日期
		if oldRequest.Date != "" {
			date, err := time.Parse("2006-01-02", oldRequest.Date)
			if err != nil {
				fmt.Printf("解析日期失败: %v，使用当前日期\n", err)
				meeting.Date = time.Now()
			} else {
				meeting.Date = date
			}
		} else {
			meeting.Date = time.Now()
		}

		// 处理待办事项
		if len(oldRequest.TodoItems) > 0 {
			meeting.TodoItems = make([]models.TodoItem, len(oldRequest.TodoItems))
			for i, item := range oldRequest.TodoItems {
				todoItem := models.TodoItem{
					ID:          uuid.New().String(),
					Description: item.Description,
					Assignee:    item.Assignee,
					Status:      item.Status,
				}

				// 解析截止日期
				if item.DueDate != "" {
					dueDate, err := time.Parse("2006-01-02", item.DueDate)
					if err == nil {
						todoItem.DueDate = dueDate
					}
				}

				meeting.TodoItems[i] = todoItem
			}
		}

		// 处理决策事项
		if len(oldRequest.Decisions) > 0 {
			meeting.Decisions = make([]models.Decision, len(oldRequest.Decisions))
			for i, item := range oldRequest.Decisions {
				meeting.Decisions[i] = models.Decision{
					ID:          uuid.New().String(),
					Description: item.Description,
					MadeBy:      item.MadeBy,
				}
			}
		}

		// 解析创建和更新时间
		if oldRequest.CreatedAt != "" {
			createdAt, err := time.Parse(time.RFC3339, oldRequest.CreatedAt)
			if err == nil {
				meeting.CreatedAt = createdAt
			} else {
				meeting.CreatedAt = time.Now()
			}
		} else {
			meeting.CreatedAt = time.Now()
		}

		if oldRequest.UpdatedAt != "" {
			updatedAt, err := time.Parse(time.RFC3339, oldRequest.UpdatedAt)
			if err == nil {
				meeting.UpdatedAt = updatedAt
			} else {
				meeting.UpdatedAt = time.Now()
			}
		} else {
			meeting.UpdatedAt = time.Now()
		}

		// 记录接收到的请求数据
		meetingBytes, _ := json.Marshal(meeting)
		fmt.Printf("转换后的会议数据：%s\n", string(meetingBytes))

		// 同步到Notion
		err = h.notionService.SyncMeeting(meeting)
		if err != nil {
			fmt.Printf("同步到Notion失败: %v\n", err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("同步到Notion失败: %v", err),
			})
		}

		return c.JSON(fiber.Map{
			"status":       "success",
			"notionPageId": meeting.NotionPageID,
		})
	}

	// 新格式请求处理
	meeting := &request.Meeting

	// 记录接收到的请求数据
	meetingBytes, _ := json.Marshal(meeting)
	fmt.Printf("【调试】解析后的会议数据：%s\n", string(meetingBytes))
	fmt.Printf("【调试】会议标题：'%s'，长度：%d\n", meeting.Title, len(meeting.Title))

	// 如果日期为空，使用当前日期
	if meeting.Date.IsZero() {
		meeting.Date = time.Now()
		fmt.Printf("日期为空，已设置为当前日期: %s\n", meeting.Date.Format("2006-01-02"))
	}

	// 确保待办事项和决策都有ID
	for i, todo := range meeting.TodoItems {
		if todo.ID == "" {
			meeting.TodoItems[i].ID = uuid.New().String()
		}
	}

	for i, decision := range meeting.Decisions {
		if decision.ID == "" {
			meeting.Decisions[i].ID = uuid.New().String()
		}
	}

	// 同步到Notion
	err := h.notionService.SyncMeeting(meeting)
	if err != nil {
		fmt.Printf("同步到Notion失败: %v\n", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("同步到Notion失败: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"status":       "success",
		"notionPageId": meeting.NotionPageID,
	})
}

// SyncToNotion 同步会议数据到Notion
func (h *Handler) SyncToNotion(c *fiber.Ctx) error {
	var request struct {
		ID           string   `json:"id"`
		Title        string   `json:"title"`
		Date         string   `json:"date"`
		Participants []string `json:"participants"`
		Transcript   string   `json:"transcript"`
		Summary      string   `json:"summary"`
		TodoItems    []struct {
			Description string `json:"description"`
			Assignee    string `json:"assignee"`
			DueDate     string `json:"dueDate"`
			Status      string `json:"status"`
		} `json:"todo_items"`
		Decisions []struct {
			Description string `json:"description"`
			MadeBy      string `json:"madeBy"`
		} `json:"decisions"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "无法解析请求数据",
		})
	}

	// 解析日期
	date, err := time.Parse("2006-01-02", request.Date)
	if err != nil {
		date = time.Now() // 如果解析失败，使用当前时间
	}

	// 构建会议对象
	meeting := &models.Meeting{
		ID:           request.ID,
		Title:        request.Title,
		Date:         date,
		Participants: request.Participants,
		Transcript:   request.Transcript,
		Summary:      request.Summary,
		TodoItems:    make([]models.TodoItem, len(request.TodoItems)),
		Decisions:    make([]models.Decision, len(request.Decisions)),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 转换待办事项
	for i, item := range request.TodoItems {
		dueDate, err := time.Parse("2006-01-02", item.DueDate)
		if err != nil {
			dueDate = time.Time{} // 如果解析失败，使用零值
		}
		meeting.TodoItems[i] = models.TodoItem{
			ID:          uuid.New().String(),
			Description: item.Description,
			Assignee:    item.Assignee,
			DueDate:     dueDate,
			Status:      item.Status,
		}
	}

	// 转换决策点
	for i, decision := range request.Decisions {
		meeting.Decisions[i] = models.Decision{
			ID:          uuid.New().String(),
			Description: decision.Description,
			MadeBy:      decision.MadeBy,
		}
	}

	// 同步到Notion
	err = h.notionService.SyncMeeting(meeting)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("同步到Notion失败: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"status":       "success",
		"notionPageId": meeting.NotionPageID,
	})
}
