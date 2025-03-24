package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"meeting-mm/config"
	"meeting-mm/models"
	"meeting-mm/services"
)

// RestHandler 处理RESTful API请求
type RestHandler struct {
	notionService *services.NotionService
}

// NewRestHandler 创建一个新的RestHandler实例
func NewRestHandler(cfg *config.Config) *RestHandler {
	return &RestHandler{
		notionService: services.NewNotionService(cfg),
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

	// 调用同步函数
	err := h.notionService.SyncMeeting(&meeting)
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
