package models

import "time"

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusProcessing TaskStatus = "processing"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed     TaskStatus = "failed"
)

type Stage struct {
	Name      string        `json:"name"`
	Status    TaskStatus    `json:"status"`
	StartTime time.Time     `json:"start_time"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
}

type Task struct {
	ID           string          `json:"id"`
	OriginalName string          `json:"original_name"`
	StoredName   string          `json:"stored_name"`
	Status       TaskStatus      `json:"status"`
	Stages       []*Stage        `json:"stages"`
	TotalTime    time.Duration   `json:"total_time"`
	Result       *AnalysisResult `json:"result,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
