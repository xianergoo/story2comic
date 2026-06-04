package model

import (
	"time"
)

type CheckpointType string

const (
	CheckpointTypeOutline     CheckpointType = "outline"
	CheckpointTypeChapterPlan CheckpointType = "chapter_plan"
	CheckpointTypeChapter     CheckpointType = "chapter"
	CheckpointTypeImage       CheckpointType = "image"
	CheckpointTypeReview      CheckpointType = "review"
)

type CheckpointStatus string

const (
	CheckpointStatusPending   CheckpointStatus = "pending"
	CheckpointStatusConfirmed CheckpointStatus = "confirmed"
	CheckpointStatusRejected  CheckpointStatus = "rejected"
	CheckpointStatusSkipped   CheckpointStatus = "skipped"
)

type Checkpoint struct {
	ID          uint            `gorm:"primarykey" json:"id"`
	AgentTaskID uint            `json:"agent_task_id"`
	NovelID     uint            `json:"novel_id"`
	ChapterNo   int             `json:"chapter_no"`
	Type        CheckpointType  `json:"type"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Data        string          `gorm:"type:text" json:"data"`
	Status      CheckpointStatus `json:"status"`
	UserAction  string          `json:"user_action"`
	UserComment string          `json:"user_comment"`
	ModifiedData string         `gorm:"type:text" json:"modified_data"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	ActionAt    *time.Time      `json:"action_at"`
}

func (Checkpoint) TableName() string {
	return "checkpoints"
}