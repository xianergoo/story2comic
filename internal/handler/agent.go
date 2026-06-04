package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"novelforge/internal/model"
	"novelforge/internal/service"
)

type AgentHandler struct {
	agentSvc *service.AgentService
	novelSvc *service.NovelService
}

func NewAgentHandler(agentSvc *service.AgentService, novelSvc *service.NovelService) *AgentHandler {
	return &AgentHandler{
		agentSvc: agentSvc,
		novelSvc: novelSvc,
	}
}

type CreateAgentTaskRequest struct {
	NovelID        uint                 `json:"novel_id" binding:"required"`
	ChapterNo      int                  `json:"chapter_no"`
	Type           model.AgentTaskType  `json:"type" binding:"required"`
	Goal           string               `json:"goal" binding:"required"`
	CheckpointMode model.CheckpointMode `json:"checkpoint_mode" binding:"required,oneof=full essential auto"`
}

func (h *AgentHandler) CreateTask(c *gin.Context) {
	var req CreateAgentTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证小说存在
	_, err := h.novelSvc.GetByID(req.NovelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "小说不存在"})
		return
	}

	task, err := h.agentSvc.CreateTask(service.CreateAgentTaskRequest{
		NovelID:        req.NovelID,
		ChapterNo:      req.ChapterNo,
		Type:           req.Type,
		Goal:           req.Goal,
		CheckpointMode: req.CheckpointMode,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (h *AgentHandler) GetTask(c *gin.Context) {
	taskID := c.Param("task_id")
	task, err := h.agentSvc.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *AgentHandler) ListTasks(c *gin.Context) {
	novelIDStr := c.Query("novel_id")
	var novelID uint = 0
	if novelIDStr != "" {
		if id, err := strconv.ParseUint(novelIDStr, 10, 32); err == nil {
			novelID = uint(id)
		}
	}

	tasks, err := h.agentSvc.ListTasks(novelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务列表失败"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *AgentHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("task_id")
	
	// 检查任务是否存在
	_, err := h.agentSvc.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	if err := h.agentSvc.CancelTask(taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消任务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务已取消"})
}

type UpdateCheckpointRequest struct {
	Action      string `json:"action" binding:"required,oneof=confirm reject skip"`
	Comment     string `json:"comment"`
	ModifiedData string `json:"modified_data"`
}

func (h *AgentHandler) UpdateCheckpoint(c *gin.Context) {
	checkpointIDStr := c.Param("checkpoint_id")
	checkpointID, err := strconv.ParseUint(checkpointIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的检查点ID"})
		return
	}

	var req UpdateCheckpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var status model.CheckpointStatus
	switch req.Action {
	case "confirm":
		status = model.CheckpointStatusConfirmed
	case "reject":
		status = model.CheckpointStatusRejected
	case "skip":
		status = model.CheckpointStatusSkipped
	}

	if err := h.agentSvc.UpdateCheckpointStatus(uint(checkpointID), status, req.Action, req.Comment, req.ModifiedData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新检查点失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "检查点已更新"})
}

func (h *AgentHandler) GetPendingCheckpoints(c *gin.Context) {
	taskID := c.Param("task_id")
	checkpoints, err := h.agentSvc.GetPendingCheckpoints(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取检查点失败"})
		return
	}
	c.JSON(http.StatusOK, checkpoints)
}