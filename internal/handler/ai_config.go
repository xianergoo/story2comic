package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	v1 "novelforge/api/v1"
	"novelforge/internal/model"
)

type AIConfigHandler struct{ db *gorm.DB }

func NewAIConfigHandler(db *gorm.DB) *AIConfigHandler { return &AIConfigHandler{db} }

type AIConfigDTO struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Provider   string `json:"provider"`
	APIKey     string `json:"api_key"`
	BaseURL    string `json:"base_url"`
	TextModel  string `json:"text_model"`
	ImageModel string `json:"image_model"`
	IsDefault  bool   `json:"is_default"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func toAIConfigDTO(cfg model.AIConfig) AIConfigDTO {
	return AIConfigDTO{
		ID:         cfg.ID,
		Name:       cfg.Name,
		Provider:   cfg.Provider,
		APIKey:     cfg.APIKey,
		BaseURL:    cfg.BaseURL,
		TextModel:  cfg.TextModel,
		ImageModel: cfg.ImageModel,
		IsDefault:  cfg.IsDefault,
		CreatedAt:  cfg.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  cfg.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (h *AIConfigHandler) Page(c *gin.Context) {
	userID := c.GetUint("user_id")
	var configs []model.AIConfig
	h.db.Where("user_id = ?", userID).Find(&configs)
	result := make([]AIConfigDTO, 0, len(configs))
	for _, cfg := range configs {
		result = append(result, toAIConfigDTO(cfg))
	}
	v1.Success(c, result)
}

type AIConfigRequest struct {
	Name       string `json:"name" binding:"required"`
	Provider   string `json:"provider" binding:"required"`
	APIKey     string `json:"api_key" binding:"required"`
	BaseURL    string `json:"base_url"`
	TextModel  string `json:"text_model" binding:"required"`
	ImageModel string `json:"image_model"`
	IsDefault  bool   `json:"is_default"`
}

func (h *AIConfigHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req AIConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.ValidationError(c, err.Error())
		return
	}

	var count int64
	h.db.Model(&model.AIConfig{}).Where("user_id = ?", userID).Count(&count)
	isDefault := req.IsDefault || count == 0
	if isDefault {
		h.db.Model(&model.AIConfig{}).Where("user_id = ?", userID).Update("is_default", false)
	}

	cfg := model.AIConfig{
		UserID:     userID,
		Name:       req.Name,
		Provider:   req.Provider,
		APIKey:     req.APIKey,
		BaseURL:    req.BaseURL,
		TextModel:  req.TextModel,
		ImageModel: req.ImageModel,
		IsDefault:  isDefault,
	}
	h.db.Create(&cfg)
	v1.SuccessWithMessage(c, "创建成功", toAIConfigDTO(cfg))
}

func (h *AIConfigHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID := c.GetUint("user_id")
	h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.AIConfig{})
	v1.Success(c, nil)
}

func (h *AIConfigHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID := c.GetUint("user_id")

	var req AIConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.ValidationError(c, err.Error())
		return
	}

	if req.IsDefault {
		h.db.Model(&model.AIConfig{}).Where("user_id = ?", userID).Update("is_default", false)
	}

	h.db.Model(&model.AIConfig{}).Where("id = ? AND user_id = ?", id, userID).Updates(map[string]interface{}{
		"name":        req.Name,
		"provider":    req.Provider,
		"api_key":     req.APIKey,
		"base_url":    req.BaseURL,
		"text_model":  req.TextModel,
		"image_model": req.ImageModel,
		"is_default":  req.IsDefault,
	})
	v1.Success(c, nil)
}
