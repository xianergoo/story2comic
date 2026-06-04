package model

import (
	"time"
)

type AgentTaskType string

const (
	AgentTaskTypeOutline     AgentTaskType = "outline"
	AgentTaskTypeChapterPlan AgentTaskType = "chapter_plan"
	AgentTaskTypeChapter     AgentTaskType = "chapter"
	AgentTaskTypeImage       AgentTaskType = "image"
	AgentTaskTypeReview      AgentTaskType = "review"
	AgentTaskTypeBatch       AgentTaskType = "batch"
)

type AgentTaskStatus string

const (
	AgentTaskStatusPending     AgentTaskStatus = "pending"
	AgentTaskStatusPlanning    AgentTaskStatus = "planning"
	AgentTaskStatusExecuting   AgentTaskStatus = "executing"
	AgentTaskStatusCheckpoint  AgentTaskStatus = "checkpoint"
	AgentTaskStatusCompleted   AgentTaskStatus = "completed"
	AgentTaskStatusFailed      AgentTaskStatus = "failed"
	AgentTaskStatusCancelled   AgentTaskStatus = "cancelled"
	AgentTaskStatusInterrupted AgentTaskStatus = "interrupted"
	AgentTaskStatusRunning     AgentTaskStatus = "running"
)

type CheckpointMode string

const (
	CheckpointModeFull      CheckpointMode = "full"
	CheckpointModeEssential CheckpointMode = "essential"
	CheckpointModeAuto      CheckpointMode = "auto"
)

type AgentTask struct {
	ID             uint            `gorm:"primarykey" json:"id"`
	TaskID         string          `gorm:"uniqueIndex;size:64" json:"task_id"`
	NovelID        uint            `json:"novel_id"`
	ChapterNo      int             `json:"chapter_no"`
	Type           AgentTaskType   `json:"type"`
	Goal           string          `json:"goal"`
	Plan           string          `gorm:"type:text" json:"plan"`
	CheckpointMode CheckpointMode  `json:"checkpoint_mode"`
	Status         AgentTaskStatus `json:"status"`
	Progress       int             `json:"progress"` // 0-100
	CurrentStep    string          `json:"current_step"`
	Result         string          `gorm:"type:text" json:"result"`
	Error          string          `gorm:"type:text" json:"error"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	CompletedAt    *time.Time      `json:"completed_at"`
	Checkpoints    []Checkpoint    `gorm:"foreignKey:AgentTaskID" json:"checkpoints"`
}

func (AgentTask) TableName() string {
	return "agent_tasks"
}