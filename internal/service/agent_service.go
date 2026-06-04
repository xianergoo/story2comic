package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"novelforge/internal/model"
)

type AgentService struct {
	db *gorm.DB
}

func NewAgentService(db *gorm.DB) *AgentService {
	return &AgentService{db: db}
}

type CreateAgentTaskRequest struct {
	NovelID        uint                    `json:"novel_id"`
	ChapterNo      int                     `json:"chapter_no"`
	Type           model.AgentTaskType     `json:"type"`
	Goal           string                  `json:"goal"`
	CheckpointMode model.CheckpointMode    `json:"checkpoint_mode"`
}

func (s *AgentService) CreateTask(req CreateAgentTaskRequest) (*model.AgentTask, error) {
	taskID := fmt.Sprintf("task_%d_%s_%d", req.NovelID, req.Type, time.Now().Unix())
	
	agentTask := &model.AgentTask{
		TaskID:         taskID,
		NovelID:        req.NovelID,
		ChapterNo:      req.ChapterNo,
		Type:           req.Type,
		Goal:           req.Goal,
		CheckpointMode: req.CheckpointMode,
		Status:         model.AgentTaskStatusPending,
		Progress:       0,
		CurrentStep:    "created",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.db.Create(agentTask).Error; err != nil {
		return nil, err
	}

	return agentTask, nil
}

func (s *AgentService) GetTask(taskID string) (*model.AgentTask, error) {
	var agentTask model.AgentTask
	if err := s.db.Preload("Checkpoints").Where("task_id = ?", taskID).First(&agentTask).Error; err != nil {
		return nil, err
	}
	return &agentTask, nil
}

func (s *AgentService) ListTasks(novelID uint) ([]model.AgentTask, error) {
	var tasks []model.AgentTask
	query := s.db.Order("created_at DESC")
	if novelID > 0 {
		query = query.Where("novel_id = ?", novelID)
	}
	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *AgentService) UpdateTaskStatus(taskID string, status model.AgentTaskStatus, progress int, step string) error {
	updates := map[string]interface{}{
		"status":       status,
		"updated_at":   time.Now(),
	}
	
	if progress >= 0 {
		updates["progress"] = progress
	}
	if step != "" {
		updates["current_step"] = step
	}
	
	if status == model.AgentTaskStatusCompleted || status == model.AgentTaskStatusFailed {
		now := time.Now()
		updates["completed_at"] = &now
	}

	return s.db.Model(&model.AgentTask{}).Where("task_id = ?", taskID).Updates(updates).Error
}

func (s *AgentService) SetTaskResult(taskID string, result string) error {
	return s.db.Model(&model.AgentTask{}).Where("task_id = ?", taskID).Updates(map[string]interface{}{
		"result":     result,
		"updated_at": time.Now(),
	}).Error
}

func (s *AgentService) SetTaskError(taskID string, errMsg string) error {
	return s.db.Model(&model.AgentTask{}).Where("task_id = ?", taskID).Updates(map[string]interface{}{
		"error":      errMsg,
		"status":     model.AgentTaskStatusFailed,
		"updated_at": time.Now(),
	}).Error
}

func (s *AgentService) CancelTask(taskID string) error {
	return s.db.Model(&model.AgentTask{}).Where("task_id = ?", taskID).Updates(map[string]interface{}{
		"status":     model.AgentTaskStatusCancelled,
		"updated_at": time.Now(),
	}).Error
}

func (s *AgentService) CreateCheckpoint(checkpoint *model.Checkpoint) error {
	checkpoint.CreatedAt = time.Now()
	checkpoint.UpdatedAt = time.Now()
	return s.db.Create(checkpoint).Error
}

func (s *AgentService) UpdateCheckpointStatus(checkpointID uint, status model.CheckpointStatus, userAction, userComment, modifiedData string) error {
	updates := map[string]interface{}{
		"status":        status,
		"updated_at":    time.Now(),
	}
	
	if userAction != "" {
		updates["user_action"] = userAction
	}
	if userComment != "" {
		updates["user_comment"] = userComment
	}
	if modifiedData != "" {
		updates["modified_data"] = modifiedData
	}
	
	if status != model.CheckpointStatusPending {
		now := time.Now()
		updates["action_at"] = &now
	}

	return s.db.Model(&model.Checkpoint{}).Where("id = ?", checkpointID).Updates(updates).Error
}

func (s *AgentService) GetPendingCheckpoints(taskID string) ([]model.Checkpoint, error) {
	var checkpoints []model.Checkpoint
	err := s.db.Joins("JOIN agent_tasks ON agent_tasks.id = checkpoints.agent_task_id").
		Where("agent_tasks.task_id = ? AND checkpoints.status = ?", taskID, model.CheckpointStatusPending).
		Find(&checkpoints).Error
	return checkpoints, err
}