package api

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	meeting := services.Meeting{
		ID:           uuid.New().String(),
		Title:        title,
		Date:         time.Now().Format("2006-01-02"),
		Participants: []string{}, // 这里可以从请求中获取参与者信息
		Transcript:   transcript,
		Summary:      summary,
		TodoItems:    todoItems,
		Decisions:    decisions,
		CreatedAt:    time.Now().Format(time.RFC3339),
		UpdatedAt:    time.Now().Format(time.RFC3339),
	}

	// 生成Markdown报告
	markdownReport, err := h.deepseekService.GenerateMarkdownReport(meeting)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("生成Markdown报告失败: %v", err),
		})
	}

	// 同步到Notion（如果需要）
	if syncToNotion {
		notionPageID, err := h.notionService.SyncToNotion(meeting, markdownReport)
		if err != nil {
			// 不中断流程，只记录错误
			fmt.Printf("同步到Notion失败: %v\n", err)
		} else {
			meeting.NotionPageID = notionPageID
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

	// 格式化当前日期为标准格式 (2006-01-02)
	currentDate := time.Now().Format("2006-01-02")
	fmt.Printf("使用的会议日期: %s\n", currentDate)

	// 创建会议对象
	meeting := services.Meeting{
		ID:           uuid.New().String(),
		Title:        title,
		Date:         currentDate,
		Participants: []string{}, // 这里可以从请求中获取参与者信息
		Transcript:   transcript,
		Summary:      summary,
		TodoItems:    todoItems,
		Decisions:    decisions,
		CreatedAt:    time.Now().Format(time.RFC3339),
		UpdatedAt:    time.Now().Format(time.RFC3339),
	}

	// 生成Markdown报告
	markdownReport, err := h.deepseekService.GenerateMarkdownReport(meeting)
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

// SyncToNotion 同步到Notion
func (h *Handler) SyncToNotion(c *fiber.Ctx) error {
	// 解析请求体
	var request struct {
		Meeting        services.Meeting `json:"meeting"`
		MarkdownReport string           `json:"markdownReport"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("解析请求体失败: %v", err),
		})
	}

	// 记录接收到的请求数据
	fmt.Printf("接收到的会议数据：%+v\n", request.Meeting)

	// 如果日期为空，使用当前日期
	if request.Meeting.Date == "" {
		request.Meeting.Date = time.Now().Format("2006-01-02")
		fmt.Printf("日期为空，已设置为当前日期: %s\n", request.Meeting.Date)
	}

	// 同步到Notion
	notionPageID, err := h.notionService.SyncToNotion(request.Meeting, request.MarkdownReport)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("同步到Notion失败: %v", err),
		})
	}

	// 返回结果
	return c.JSON(fiber.Map{
		"notionPageId": notionPageID,
	})
}
