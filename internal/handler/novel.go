package handler

import (
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
	c.HTML(http.StatusOK, "home.html", gin.H{"novels": novels})
}

func (h *NovelHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")
	title := c.PostForm("title")
	mode := c.PostForm("mode")
	imageMode := c.PostForm("image_mode")
	summary := c.PostForm("summary")
	aiConfigID, _ := strconv.Atoi(c.PostForm("ai_config_id"))
	novel, err := h.svc.Create(userID, title, summary, mode, imageMode, uint(aiConfigID))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusFound, "/novel/"+strconv.Itoa(int(novel.ID)))
}

func (h *NovelHandler) StartGeneration(c *gin.Context) {
	novelID, _ := strconv.Atoi(c.PostForm("novel_id"))
	go h.svc.StartGeneration(uint(novelID))
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
	c.HTML(http.StatusOK, "novel_detail.html", gin.H{
		"novel":    novel,
		"chapters": chapters,
		"pages":    pages,
	})
}
