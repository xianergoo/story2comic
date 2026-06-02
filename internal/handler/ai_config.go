package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"novelforge/internal/model"
)

type AIConfigHandler struct{ db *gorm.DB }

func NewAIConfigHandler(db *gorm.DB) *AIConfigHandler { return &AIConfigHandler{db} }

func (h *AIConfigHandler) Page(c *gin.Context) {
	userID := c.GetUint("user_id")
	var configs []model.AIConfig
	h.db.Where("user_id = ?", userID).Find(&configs)
	c.HTML(http.StatusOK, "ai_config.html", gin.H{"configs": configs})
}

func (h *AIConfigHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")

	// 第一个配置自动设为默认
	var count int64
	h.db.Model(&model.AIConfig{}).Where("user_id = ?", userID).Count(&count)
	isDefault := c.PostForm("is_default") == "on" || count == 0
	if isDefault {
		h.db.Model(&model.AIConfig{}).Where("user_id = ?", userID).Update("is_default", false)
	}
	cfg := model.AIConfig{
		UserID:     userID,
		Name:       c.PostForm("name"),
		Provider:   c.PostForm("provider"),
		APIKey:     c.PostForm("api_key"),
		BaseURL:    c.PostForm("base_url"),
		TextModel:  c.PostForm("text_model"),
		ImageModel: c.PostForm("image_model"),
		IsDefault:  isDefault,
	}
	h.db.Create(&cfg)
	c.Redirect(http.StatusFound, "/ai-config")
}

func (h *AIConfigHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID := c.GetUint("user_id")
	h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.AIConfig{})
	c.Redirect(http.StatusFound, "/ai-config")
}

func (h *AIConfigHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID := c.GetUint("user_id")

	isDefault := c.PostForm("is_default") == "on"
	if isDefault {
		h.db.Model(&model.AIConfig{}).Where("user_id = ?", userID).Update("is_default", false)
	}

	h.db.Model(&model.AIConfig{}).Where("id = ? AND user_id = ?", id, userID).Updates(map[string]interface{}{
		"name":        c.PostForm("name"),
		"provider":    c.PostForm("provider"),
		"api_key":     c.PostForm("api_key"),
		"base_url":    c.PostForm("base_url"),
		"text_model":  c.PostForm("text_model"),
		"image_model": c.PostForm("image_model"),
		"is_default":  isDefault,
	})
	c.Redirect(http.StatusFound, "/ai-config")
}
