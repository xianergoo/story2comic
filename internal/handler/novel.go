package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"novelforge/internal/model"
	"novelforge/internal/service"
)

type NovelHandler struct {
	svc *service.NovelService
	db  *gorm.DB
}

func NewNovelHandler(svc *service.NovelService, db *gorm.DB) *NovelHandler {
	return &NovelHandler{svc: svc, db: db}
}

func (h *NovelHandler) Home(c *gin.Context) {
	userID := c.GetUint("user_id")
	novels, _ := h.svc.List(userID)
	type NovelWithProgress struct {
		model.Novel
		ChapterCount int64
	}
	var result []NovelWithProgress
	for _, n := range novels {
		var count int64
		h.db.Model(&model.Chapter{}).Where("novel_id = ? AND status = ?", n.ID, "done").Count(&count)
		result = append(result, NovelWithProgress{Novel: n, ChapterCount: count})
	}
	c.HTML(http.StatusOK, "home.html", gin.H{"novels": result})
}

func (h *NovelHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")
	title := c.PostForm("title")
	mode := c.PostForm("mode")
	imageMode := c.PostForm("image_mode")
	summary := c.PostForm("summary")
	aiConfigID, _ := strconv.Atoi(c.PostForm("ai_config_id"))
	if aiConfigID == 0 {
		var defaultCfg model.AIConfig
		err := h.db.Where("user_id = ? AND is_default = ?", userID, true).First(&defaultCfg).Error
		if err != nil {
			err = h.db.Where("user_id = ?", userID).First(&defaultCfg).Error
		}
		if err != nil {
			c.HTML(http.StatusBadRequest, "home.html", gin.H{"novels": nil, "error": "请先在 AI配置页面 添加 API Key 配置"})
			return
		}
		aiConfigID = int(defaultCfg.ID)
	}
	novel, err := h.svc.Create(userID, title, summary, mode, imageMode, uint(aiConfigID))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	go func() {
		if err := h.svc.StartGeneration(novel.ID); err != nil {
			fmt.Println("StartGeneration failed:", err.Error())
		}
	}()
	c.Redirect(http.StatusFound, "/novel/"+strconv.Itoa(int(novel.ID)))
}

func (h *NovelHandler) StartGeneration(c *gin.Context) {
	novelID, _ := strconv.Atoi(c.PostForm("novel_id"))
	go h.svc.StartGeneration(uint(novelID))
	c.Redirect(http.StatusFound, "/novel/"+strconv.Itoa(novelID))
}

func (h *NovelHandler) Stop(c *gin.Context) {
	novelID, _ := strconv.Atoi(c.Param("id"))
	pipeline := c.DefaultQuery("pipeline", "text")
	if pipeline == "image" {
		h.db.Model(&model.Novel{}).Where("id = ?", novelID).Update("image_status", "paused")
	} else {
		h.db.Model(&model.Novel{}).Where("id = ?", novelID).Updates(map[string]any{
			"text_status": "paused", "status": "stopped",
		})
	}
	c.Redirect(http.StatusFound, "/novel/"+strconv.Itoa(novelID))
}

func (h *NovelHandler) Resume(c *gin.Context) {
	novelID, _ := strconv.Atoi(c.Param("id"))
	pipeline := c.DefaultQuery("pipeline", "text")
	if pipeline == "image" {
		h.db.Model(&model.Novel{}).Where("id = ?", novelID).Update("image_status", "generating")
	} else {
		h.db.Model(&model.Novel{}).Where("id = ?", novelID).Updates(map[string]any{
			"text_status": "writing", "status": "drafting",
		})
		go h.svc.Resume(uint(novelID))
	}
	c.Redirect(http.StatusFound, "/novel/"+strconv.Itoa(novelID))
}

func (h *NovelHandler) Detail(c *gin.Context) {
	novelID, _ := strconv.Atoi(c.Param("id"))
	novel, err := h.svc.GetByID(uint(novelID))
	if err != nil {
		c.String(http.StatusNotFound, "作品不存在")
		return
	}
	var chapters []model.Chapter
	h.db.Where("novel_id = ?", novelID).Order("chapter_no").Find(&chapters)
	var pages []model.ComicPage
	h.db.Where("novel_id = ?", novelID).Order("chapter_id, page_no").Find(&pages)
	var outline model.Outline
	h.db.Where("novel_id = ?", novelID).Order("version DESC").First(&outline)

	var textDone, comicDone int
	for _, ch := range chapters {
		if ch.Status == "done" {
			textDone++
		}
	}
	for _, p := range pages {
		if p.Status == "done" {
			comicDone++
		}
	}

	c.HTML(http.StatusOK, "novel_detail.html", gin.H{
		"novel":     novel,
		"chapters":  chapters,
		"pages":     pages,
		"outline":   outline,
		"textDone":  textDone,
		"comicDone": comicDone,
	})
}
