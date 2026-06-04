package handler

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	v1 "novelforge/api/v1"
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

type NovelListItem struct {
	ID           uint   `json:"id"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	Mode         string `json:"mode"`
	ImageMode    string `json:"image_mode"`
	Status       string `json:"status"`
	TextStatus   string `json:"text_status"`
	ImageStatus  string `json:"image_status"`
	ChapterCount int64  `json:"chapter_count"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type NovelDTO struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"user_id"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	CoverURL    string `json:"cover_url"`
	Mode        string `json:"mode"`
	ImageMode   string `json:"image_mode"`
	Status      string `json:"status"`
	TextStatus  string `json:"text_status"`
	ImageStatus string `json:"image_status"`
	AIConfigID  uint   `json:"ai_config_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ChapterDTO struct {
	ID           uint   `json:"id"`
	NovelID      uint   `json:"novel_id"`
	ChapterNo    int    `json:"chapter_no"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Status       string `json:"status"`
	RewriteCount int    `json:"rewrite_count"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type OutlineDTO struct {
	ID              uint   `json:"id"`
	NovelID         uint   `json:"novel_id"`
	Version         int    `json:"version"`
	Content         string `json:"content"`
	CharacterSheets string `json:"character_sheets"`
	WorldSetting    string `json:"world_setting"`
	ChapterPlan     string `json:"chapter_plan"`
	CreatedAt       string `json:"created_at"`
}

type ComicPageDTO struct {
	ID         uint   `json:"id"`
	ChapterID  uint   `json:"chapter_id"`
	NovelID    uint   `json:"novel_id"`
	PageNo     int    `json:"page_no"`
	PanelCount int    `json:"panel_count"`
	ImageURLs  string `json:"image_urls"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func toNovelDTO(n *model.Novel) NovelDTO {
	return NovelDTO{ID: n.ID, UserID: n.UserID, Title: n.Title, Summary: n.Summary, CoverURL: n.CoverURL, Mode: n.Mode, ImageMode: n.ImageMode, Status: n.Status, TextStatus: n.TextStatus, ImageStatus: n.ImageStatus, AIConfigID: n.AIConfigID, CreatedAt: n.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: n.UpdatedAt.Format("2006-01-02 15:04:05")}
}

func toChapterDTO(ch model.Chapter) ChapterDTO {
	return ChapterDTO{ID: ch.ID, NovelID: ch.NovelID, ChapterNo: ch.ChapterNo, Title: ch.Title, Content: ch.Content, Status: ch.Status, RewriteCount: ch.RewriteCount, CreatedAt: ch.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: ch.UpdatedAt.Format("2006-01-02 15:04:05")}
}

func toOutlineDTO(o model.Outline) OutlineDTO {
	return OutlineDTO{ID: o.ID, NovelID: o.NovelID, Version: o.Version, Content: o.Content, CharacterSheets: o.CharacterSheets, WorldSetting: o.WorldSetting, ChapterPlan: o.ChapterPlan, CreatedAt: o.CreatedAt.Format("2006-01-02 15:04:05")}
}

func toComicPageDTO(p model.ComicPage) ComicPageDTO {
	return ComicPageDTO{ID: p.ID, ChapterID: p.ChapterID, NovelID: p.NovelID, PageNo: p.PageNo, PanelCount: p.PanelCount, ImageURLs: p.ImageURLs, Status: p.Status, CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: p.UpdatedAt.Format("2006-01-02 15:04:05")}
}

func (h *NovelHandler) Home(c *gin.Context) {
	userID := c.GetUint("user_id")
	novels, _ := h.svc.List(userID)

	var result []NovelListItem
	for _, n := range novels {
		var count int64
		h.db.Model(&model.Chapter{}).Where("novel_id = ? AND status = ?", n.ID, "done").Count(&count)
		result = append(result, NovelListItem{
			ID:           n.ID,
			Title:        n.Title,
			Summary:      n.Summary,
			Mode:         n.Mode,
			ImageMode:    n.ImageMode,
			Status:       n.Status,
			TextStatus:   n.TextStatus,
			ImageStatus:  n.ImageStatus,
			ChapterCount: count,
			CreatedAt:    n.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    n.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	v1.Success(c, result)
}

type CreateNovelRequest struct {
	Title      string `json:"title" binding:"required"`
	Summary    string `json:"summary"`
	Mode       string `json:"mode" binding:"required,oneof=inspiration outline blindbox"`
	ImageMode  string `json:"image_mode"`
	AIConfigID uint   `json:"ai_config_id"`
}

func (h *NovelHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateNovelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.ValidationError(c, err.Error())
		return
	}
	if req.Mode == "" {
		req.Mode = "inspiration"
	}
	if req.ImageMode == "" {
		req.ImageMode = "single"
	}

	aiConfigID := req.AIConfigID
	if aiConfigID == 0 {
		var defaultCfg model.AIConfig
		err := h.db.Where("user_id = ? AND is_default = ?", userID, true).First(&defaultCfg).Error
		if err != nil {
			err = h.db.Where("user_id = ?", userID).First(&defaultCfg).Error
		}
		if err != nil {
			v1.Error(c, 400, "请先添加 AI 配置")
			return
		}
		aiConfigID = defaultCfg.ID
	}

	novel, err := h.svc.Create(userID, req.Title, req.Summary, req.Mode, req.ImageMode, aiConfigID)
	if err != nil {
		v1.InternalServerError(c, err.Error())
		return
	}

	go func() {
		if err := h.svc.StartGeneration(novel.ID); err != nil {
			fmt.Println("StartGeneration failed:", err.Error())
		}
	}()

	v1.SuccessWithMessage(c, "创建成功", gin.H{"id": novel.ID, "title": novel.Title})
}

type StopResumeRequest struct {
	Pipeline string `json:"pipeline"`
}

func (h *NovelHandler) Stop(c *gin.Context) {
	novelID, _ := strconv.Atoi(c.Param("id"))

	var req StopResumeRequest
	c.ShouldBindJSON(&req)
	if req.Pipeline == "" {
		req.Pipeline = "text"
	}

	if req.Pipeline == "image" {
		h.db.Model(&model.Novel{}).Where("id = ?", novelID).Update("image_status", "paused")
	} else {
		h.db.Model(&model.Novel{}).Where("id = ?", novelID).Updates(map[string]any{
			"text_status": "paused", "status": "stopped",
		})
	}
	v1.Success(c, gin.H{"id": novelID, "pipeline": req.Pipeline})
}

func (h *NovelHandler) Resume(c *gin.Context) {
	novelID, _ := strconv.Atoi(c.Param("id"))

	var req StopResumeRequest
	c.ShouldBindJSON(&req)
	if req.Pipeline == "" {
		req.Pipeline = "text"
	}

	if req.Pipeline == "image" {
		h.db.Model(&model.Novel{}).Where("id = ?", novelID).Update("image_status", "generating")
	} else {
		h.db.Model(&model.Novel{}).Where("id = ?", novelID).Updates(map[string]any{
			"text_status": "writing", "status": "drafting",
		})
		go h.svc.Resume(uint(novelID))
	}
	v1.Success(c, gin.H{"id": novelID, "pipeline": req.Pipeline})
}

type NovelDetailResponse struct {
	Novel    NovelDTO        `json:"novel"`
	Chapters []ChapterDTO    `json:"chapters"`
	Pages    []ComicPageDTO  `json:"pages"`
	Outline  *OutlineDTO     `json:"outline"`
	Progress struct {
		Planned   int `json:"planned"`
		TextDone  int `json:"text_done"`
		ComicDone int `json:"comic_done"`
		NextNo    int `json:"next_chapter_no"`
	} `json:"progress"`
}

func (h *NovelHandler) Detail(c *gin.Context) {
	novelID, _ := strconv.Atoi(c.Param("id"))
	novel, err := h.svc.GetByID(uint(novelID))
	if err != nil {
		v1.NotFound(c, "作品不存在")
		return
	}

	var chapters []model.Chapter
	h.db.Where("novel_id = ?", novelID).Order("chapter_no").Find(&chapters)

	var pages []model.ComicPage
	h.db.Where("novel_id = ?", novelID).Order("chapter_id, page_no").Find(&pages)

	var outline model.Outline
	h.db.Where("novel_id = ?", novelID).Order("version DESC").First(&outline)

	var textDone, planned, nextNo int
	for _, ch := range chapters {
		if ch.Status == "done" {
			textDone++
		}
	}

	var chapterPlan []service.ChapterPlanItem
	json.Unmarshal([]byte(outline.ChapterPlan), &chapterPlan)
	planned = len(chapterPlan)

	for _, cp := range chapterPlan {
		found := false
		for _, ch := range chapters {
			if ch.ChapterNo == cp.ChapterNo && ch.Status == "done" {
				found = true
				break
			}
		}
		if !found {
			nextNo = cp.ChapterNo
			break
		}
	}

	var comicDone int
	for _, p := range pages {
		if p.Status == "done" {
			comicDone++
		}
	}

	chapterDTOs := make([]ChapterDTO, 0, len(chapters))
	for _, ch := range chapters {
		chapterDTOs = append(chapterDTOs, toChapterDTO(ch))
	}
	pageDTOs := make([]ComicPageDTO, 0, len(pages))
	for _, p := range pages {
		pageDTOs = append(pageDTOs, toComicPageDTO(p))
	}
	outlineDTO := toOutlineDTO(outline)
	resp := NovelDetailResponse{
		Novel:    toNovelDTO(novel),
		Chapters: chapterDTOs,
		Pages:    pageDTOs,
		Outline:  &outlineDTO,
	}
	resp.Progress.Planned = planned
	resp.Progress.TextDone = textDone
	resp.Progress.ComicDone = comicDone
	resp.Progress.NextNo = nextNo

	v1.Success(c, resp)
}

type RegenerateRequest struct {
	Suggestion string `json:"suggestion"`
}

func (h *NovelHandler) RegenerateChapter(c *gin.Context) {
	novelID, _ := strconv.Atoi(c.Param("id"))
	chapterNo, _ := strconv.Atoi(c.Param("no"))

	var req RegenerateRequest
	c.ShouldBindJSON(&req)

	h.db.Where("novel_id = ? AND chapter_no = ?", novelID, chapterNo).Delete(&model.Chapter{})
	h.svc.EnqueueChapterWithSuggestion(uint(novelID), chapterNo, req.Suggestion)

	v1.Success(c, gin.H{"novel_id": novelID, "chapter_no": chapterNo})
}
