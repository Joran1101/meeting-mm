package models

import (
	"time"
)

// Meeting 表示一个会议记录
type Meeting struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Date         time.Time  `json:"date"`
	Participants []string   `json:"participants"`
	Transcript   string     `json:"transcript"`
	Summary      string     `json:"summary"`
	TodoItems    []TodoItem `json:"todoItems"`
	Decisions    []Decision `json:"decisions"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	NotionPageID string     `json:"notionPageId,omitempty"`
}

// TodoItem 表示从会议中提取的待办事项
type TodoItem struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Assignee    string    `json:"assignee"`
	DueDate     time.Time `json:"dueDate,omitempty"`
	Status      string    `json:"status"` // "pending", "completed", "in_progress"
}

// Decision 表示从会议中提取的决策点
type Decision struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	MadeBy      string `json:"madeBy,omitempty"`
}

// TranscriptSegment 表示语音转文字的一个片段
type TranscriptSegment struct {
	ID        string    `json:"id"`
	MeetingID string    `json:"meetingId"`
	StartTime float64   `json:"startTime"` // 以秒为单位
	EndTime   float64   `json:"endTime"`   // 以秒为单位
	Speaker   string    `json:"speaker,omitempty"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}
