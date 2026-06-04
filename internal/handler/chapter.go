package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	v1 "novelforge/api/v1"
	"novelforge/internal/model"
	"novelforge/internal/service"
)

type ViewComicPage struct {
	PageNo     int      `json:"page_no"`
	PanelCount int      `json:"panel_count"`
	ImageURLs  []string `json:"image_urls"`
}

type ChapterHandler struct {
	db  *gorm.DB
	nSvc *service.NovelService
	cSvc *service.ChapterService
}

func NewChapterHandler(db *gorm.DB, nSvc *service.NovelService, cSvc *service.ChapterService) *ChapterHandler {
	return &ChapterHandler{db: db, nSvc: nSvc, cSvc: cSvc}
}

type ChapterViewResponse struct {
	Novel   NovelDTO        `json:"novel"`
	Chapter ChapterDTO      `json:"chapter"`
	Pages   []ViewComicPage `json:"pages"`
	PrevNo  int             `json:"prev_chapter_no"`
	NextNo  int             `json:"next_chapter_no"`
}

func (h *ChapterHandler) View(c *gin.Context) {
	novelID, _ := strconv.Atoi(c.Param("id"))
	chapterNo, _ := strconv.Atoi(c.Param("no"))

	novel, err := h.nSvc.GetByID(uint(novelID))
	if err != nil {
		v1.NotFound(c, "作品不存在")
		return
	}

	chapter, err := h.cSvc.GetByNovelAndNo(uint(novelID), chapterNo)
	if err != nil {
		v1.NotFound(c, "章节不存在")
		return
	}

	var dbPages []model.ComicPage
	h.db.Where("novel_id = ? AND chapter_id = ?", novelID, chapter.ID).Order("page_no").Find(&dbPages)

	var viewPages []ViewComicPage
	for _, p := range dbPages {
		var urls []string
		json.Unmarshal([]byte(p.ImageURLs), &urls)
		viewPages = append(viewPages, ViewComicPage{
			PageNo:     p.PageNo,
			PanelCount: p.PanelCount,
			ImageURLs:  urls,
		})
	}

	var prevNo, nextNo int
	h.db.Model(&model.Chapter{}).
		Where("novel_id = ? AND chapter_no < ?", novelID, chapterNo).
		Select("COALESCE(MAX(chapter_no), 0)").Scan(&prevNo)
	h.db.Model(&model.Chapter{}).
		Where("novel_id = ? AND chapter_no > ?", novelID, chapterNo).
		Select("COALESCE(MIN(chapter_no), 0)").Scan(&nextNo)

	v1.Success(c, ChapterViewResponse{
		Novel:   toNovelDTO(novel),
		Chapter: toChapterDTO(*chapter),
		Pages:   viewPages,
		PrevNo:  prevNo,
		NextNo:  nextNo,
	})
}
