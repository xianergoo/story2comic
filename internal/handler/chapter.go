package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"novelforge/internal/model"
	"novelforge/internal/service"
)

type ViewComicPage struct {
	PanelCount int
	ImageURLs  []string
}

type ChapterHandler struct {
	db   *gorm.DB
	nSvc *service.NovelService
	cSvc *service.ChapterService
}

func NewChapterHandler(db *gorm.DB, nSvc *service.NovelService, cSvc *service.ChapterService) *ChapterHandler {
	return &ChapterHandler{db: db, nSvc: nSvc, cSvc: cSvc}
}

func (h *ChapterHandler) View(c *gin.Context) {
	novelID, _ := strconv.Atoi(c.Param("id"))
	chapterNo, _ := strconv.Atoi(c.Param("no"))

	novel, err := h.nSvc.GetByID(uint(novelID))
	if err != nil {
		c.String(http.StatusNotFound, "作品不存在")
		return
	}

	chapter, err := h.cSvc.GetByNovelAndNo(uint(novelID), chapterNo)
	if err != nil {
		c.String(http.StatusNotFound, "章节不存在")
		return
	}

	var dbPages []model.ComicPage
	h.db.Where("novel_id = ? AND chapter_id = ?", novelID, chapter.ID).Order("page_no").Find(&dbPages)

	var viewPages []ViewComicPage
	for _, p := range dbPages {
		var urls []string
		json.Unmarshal([]byte(p.ImageURLs), &urls)
		viewPages = append(viewPages, ViewComicPage{PanelCount: p.PanelCount, ImageURLs: urls})
	}

	var prevNo, nextNo int
	h.db.Model(&model.Chapter{}).
		Where("novel_id = ? AND chapter_no < ?", novelID, chapterNo).
		Select("COALESCE(MAX(chapter_no), 0)").Scan(&prevNo)
	h.db.Model(&model.Chapter{}).
		Where("novel_id = ? AND chapter_no > ?", novelID, chapterNo).
		Select("COALESCE(MIN(chapter_no), 0)").Scan(&nextNo)

	c.HTML(http.StatusOK, "reader.html", gin.H{
		"novel":   novel,
		"chapter": chapter,
		"pages":   viewPages,
		"prevNo":  prevNo,
		"nextNo":  nextNo,
	})
}
