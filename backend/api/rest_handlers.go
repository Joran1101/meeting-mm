package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"meeting-mm/models"
	"meeting-mm/services"
)

// RestHandler 处理RESTful API请求
type RestHandler struct {
	syncNotionService func(title string, date time.Time, summary string, todos []string, decisions []string) error
}

// NewRestHandler 创建一个新的RestHandler实例
func NewRestHandler() *RestHandler {
	return &RestHandler{
		syncNotionService: services.SyncToNotion,
	}
}

// HealthCheckHandler 处理健康检查请求
func (h *RestHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}
	respondWithJSON(w, http.StatusOK, response)
}

// SyncMeetingToNotionHandler 处理将会议同步到Notion的请求
func (h *RestHandler) SyncMeetingToNotionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	var meeting models.Meeting
	if err := json.NewDecoder(r.Body).Decode(&meeting); err != nil {
		respondWithError(w, http.StatusBadRequest, "无效的请求格式")
		return
	}

	// 打印接收到的会议数据
	meetingBytes, _ := json.Marshal(meeting)
	fmt.Printf("接收到的会议数据: %s\n", string(meetingBytes))

	// 获取会议日期，已经是time.Time类型
	meetingDate := meeting.Date

	// 如果日期为零值，使用当前日期
	if meetingDate.IsZero() {
		fmt.Printf("日期为零值，使用当前日期\n")
		meetingDate = time.Now()
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

	// 调用同步函数
	err := h.syncNotionService(
		meeting.Title,
		meetingDate,
		meeting.Summary,
		todos,
		decisions,
	)

	if err != nil {
		fmt.Printf("同步到Notion失败: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("同步到Notion失败: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "会议纪要已成功同步到Notion",
	})
}

