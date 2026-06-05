package handler

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	v1 "novelforge/api/v1"
	"novelforge/internal/model"
	"novelforge/internal/service"
)

type ChapterPlanSnapshot struct {
	ChapterNo int    `json:"chapter_no"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
}

type NovelOutlineSnapshot struct {
	ID              uint                              `json:"id"`
	Version         int                               `json:"version"`
	Content         string                            `json:"content"`
	WorldSetting    string                            `json:"world_setting"`
	CharacterSheets map[string]service.CharacterSheet `json:"character_sheets"`
	ChapterPlan     []ChapterPlanSnapshot             `json:"chapter_plan"`
	CreatedAt       string                            `json:"created_at"`
}

type NovelChapterComicSnapshot struct {
	PageCount     int    `json:"page_count"`
	DonePageCount int    `json:"done_page_count"`
	Status        string `json:"status"`
}

type NovelChapterSnapshot struct {
	ChapterNo    int                       `json:"chapter_no"`
	Title        string                    `json:"title"`
	Summary      string                    `json:"summary"`
	State        string                    `json:"state"`
	Status       string                    `json:"status"`
	HasContent   bool                      `json:"has_content"`
	RewriteCount int                       `json:"rewrite_count"`
	UpdatedAt    string                    `json:"updated_at"`
	Comic        NovelChapterComicSnapshot `json:"comic"`
}

type NovelProgressSnapshot struct {
	PlannedCount     int `json:"planned_count"`
	GeneratedCount   int `json:"generated_count"`
	TextDoneCount    int `json:"text_done_count"`
	RemainingCount   int `json:"remaining_count"`
	CurrentChapterNo int `json:"current_chapter_no"`
}

type NovelFocusSnapshot struct {
	ChapterNo int    `json:"chapter_no"`
	Title     string `json:"title"`
	State     string `json:"state"`
	Status    string `json:"status"`
}

type NovelComicSummarySnapshot struct {
	ChapterCount        int `json:"chapter_count"`
	ChapterDoneCount    int `json:"chapter_done_count"`
	PageCount           int `json:"page_count"`
	DonePageCount       int `json:"done_page_count"`
	PendingChapterCount int `json:"pending_chapter_count"`
}

type comicAggregate struct {
	pageCount     int
	donePageCount int
}

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

func parseChapterPlanSnapshots(raw string) []ChapterPlanSnapshot {
	var plan []service.ChapterPlanItem
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return []ChapterPlanSnapshot{}
	}
	result := make([]ChapterPlanSnapshot, 0, len(plan))
	for _, item := range plan {
		result = append(result, ChapterPlanSnapshot{
			ChapterNo: item.ChapterNo,
			Title:     item.Title,
			Summary:   item.Summary,
		})
	}
	return result
}

func parseCharacterSheets(raw string) map[string]service.CharacterSheet {
	if raw == "" {
		return map[string]service.CharacterSheet{}
	}
	var sheets map[string]service.CharacterSheet
	if err := json.Unmarshal([]byte(raw), &sheets); err != nil {
		return map[string]service.CharacterSheet{}
	}
	if sheets == nil {
		return map[string]service.CharacterSheet{}
	}
	return sheets
}

func parseImageURLs(raw string) []string {
	if raw == "" {
		return []string{}
	}
	var urls []string
	if err := json.Unmarshal([]byte(raw), &urls); err != nil {
		return []string{}
	}
	if urls == nil {
		return []string{}
	}
	return urls
}

func toOutlineSnapshot(o *model.Outline) *NovelOutlineSnapshot {
	if o == nil {
		return nil
	}
	return &NovelOutlineSnapshot{
		ID:              o.ID,
		Version:         o.Version,
		Content:         o.Content,
		WorldSetting:    o.WorldSetting,
		CharacterSheets: parseCharacterSheets(o.CharacterSheets),
		ChapterPlan:     parseChapterPlanSnapshots(o.ChapterPlan),
		CreatedAt:       o.CreatedAt.Format("2006-01-02 15:04:05"),
	}
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
	Novel    NovelDTO                  `json:"novel"`
	Chapters []NovelChapterSnapshot    `json:"chapters"`
	Outline  *NovelOutlineSnapshot     `json:"outline"`
	Progress NovelProgressSnapshot     `json:"progress"`
	Focus    *NovelFocusSnapshot       `json:"focus"`
	Comic    NovelComicSummarySnapshot `json:"comic"`
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
	outlineErr := h.db.Where("novel_id = ?", novelID).Order("version DESC").First(&outline).Error

	var outlineModel *model.Outline
	if outlineErr == nil {
		outlineModel = &outline
	}

	plan := parseChapterPlanSnapshots("")
	if outlineModel != nil {
		plan = parseChapterPlanSnapshots(outlineModel.ChapterPlan)
	}

	chaptersByNo := make(map[int]model.Chapter, len(chapters))
	generatedCount := 0
	textDoneCount := 0
	for _, ch := range chapters {
		chaptersByNo[ch.ChapterNo] = ch
		generatedCount++
		if ch.Status == "done" {
			textDoneCount++
		}
	}

	comicByChapterNo := make(map[int]comicAggregate, len(chapters))
	totalPageCount := 0
	donePageCount := 0
	chapterNoByID := make(map[uint]int, len(chapters))
	for _, ch := range chapters {
		chapterNoByID[ch.ID] = ch.ChapterNo
	}
	for _, p := range pages {
		chapterNo, ok := chapterNoByID[p.ChapterID]
		if !ok {
			continue
		}
		agg := comicByChapterNo[chapterNo]
		agg.pageCount++
		totalPageCount++
		if p.Status == "done" {
			agg.donePageCount++
			donePageCount++
		}
		comicByChapterNo[chapterNo] = agg
	}

	chapterSnapshots := make([]NovelChapterSnapshot, 0, len(plan)+len(chapters))
	planChapterNos := make(map[int]bool, len(plan))
	var focus *NovelFocusSnapshot

	for _, item := range plan {
		planChapterNos[item.ChapterNo] = true
		snapshot := NovelChapterSnapshot{
			ChapterNo: item.ChapterNo,
			Title:     item.Title,
			Summary:   item.Summary,
			State:     "planned",
			Status:    "planned",
			Comic: NovelChapterComicSnapshot{
				Status: "planned",
			},
		}
		if ch, ok := chaptersByNo[item.ChapterNo]; ok {
			snapshot.State = "generated"
			snapshot.Status = ch.Status
			snapshot.HasContent = ch.Content != ""
			snapshot.RewriteCount = ch.RewriteCount
			snapshot.UpdatedAt = ch.UpdatedAt.Format("2006-01-02 15:04:05")
		}
		if comic, ok := comicByChapterNo[item.ChapterNo]; ok {
			snapshot.Comic.PageCount = comic.pageCount
			snapshot.Comic.DonePageCount = comic.donePageCount
			snapshot.Comic.Status = "pending"
			if comic.pageCount > 0 && comic.pageCount == comic.donePageCount {
				snapshot.Comic.Status = "done"
			}
		}
		chapterSnapshots = append(chapterSnapshots, snapshot)
	}

	for _, ch := range chapters {
		if planChapterNos[ch.ChapterNo] {
			continue
		}
		snapshot := NovelChapterSnapshot{
			ChapterNo:    ch.ChapterNo,
			Title:        ch.Title,
			State:        "generated",
			Status:       ch.Status,
			HasContent:   ch.Content != "",
			RewriteCount: ch.RewriteCount,
			UpdatedAt:    ch.UpdatedAt.Format("2006-01-02 15:04:05"),
			Comic: NovelChapterComicSnapshot{
				Status: "planned",
			},
		}
		if comic, ok := comicByChapterNo[ch.ChapterNo]; ok {
			snapshot.Comic.PageCount = comic.pageCount
			snapshot.Comic.DonePageCount = comic.donePageCount
			snapshot.Comic.Status = "pending"
			if comic.pageCount > 0 && comic.pageCount == comic.donePageCount {
				snapshot.Comic.Status = "done"
			}
		}
		chapterSnapshots = append(chapterSnapshots, snapshot)
	}

	sort.Slice(chapterSnapshots, func(i, j int) bool {
		return chapterSnapshots[i].ChapterNo < chapterSnapshots[j].ChapterNo
	})

	remainingCount := 0
	currentChapterNo := 0
	if len(chapterSnapshots) > 0 {
		focus = nil
		for _, snapshot := range chapterSnapshots {
			if planChapterNos[snapshot.ChapterNo] && snapshot.Status != "done" {
				remainingCount++
				if currentChapterNo == 0 {
					currentChapterNo = snapshot.ChapterNo
				}
			}
			if snapshot.Status == "done" {
				continue
			}
			if focus == nil {
				focus = &NovelFocusSnapshot{
					ChapterNo: snapshot.ChapterNo,
					Title:     snapshot.Title,
					State:     snapshot.State,
					Status:    snapshot.Status,
				}
			}
		}
		if focus == nil {
			last := chapterSnapshots[len(chapterSnapshots)-1]
			focus = &NovelFocusSnapshot{
				ChapterNo: last.ChapterNo,
				Title:     last.Title,
				State:     last.State,
				Status:    last.Status,
			}
		}
	}

	comicChapterDoneCount := 0
	comicPendingChapterCount := 0
	for _, snapshot := range chapterSnapshots {
		if snapshot.Comic.Status == "done" {
			comicChapterDoneCount++
			continue
		}
		comicPendingChapterCount++
	}

	resp := NovelDetailResponse{
		Novel:    toNovelDTO(novel),
		Chapters: chapterSnapshots,
		Outline:  toOutlineSnapshot(outlineModel),
		Progress: NovelProgressSnapshot{
			PlannedCount:     len(plan),
			GeneratedCount:   generatedCount,
			TextDoneCount:    textDoneCount,
			RemainingCount:   remainingCount,
			CurrentChapterNo: currentChapterNo,
		},
		Focus: focus,
		Comic: NovelComicSummarySnapshot{
			ChapterCount:        len(chapterSnapshots),
			ChapterDoneCount:    comicChapterDoneCount,
			PageCount:           totalPageCount,
			DonePageCount:       donePageCount,
			PendingChapterCount: comicPendingChapterCount,
		},
	}

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
