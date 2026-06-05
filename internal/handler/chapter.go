package handler

import (
	"sort"
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
	db   *gorm.DB
	nSvc *service.NovelService
	cSvc *service.ChapterService
}

func NewChapterHandler(db *gorm.DB, nSvc *service.NovelService, cSvc *service.ChapterService) *ChapterHandler {
	return &ChapterHandler{db: db, nSvc: nSvc, cSvc: cSvc}
}

type ChapterViewResponse struct {
	Novel       NovelDTO              `json:"novel"`
	State       string                `json:"state"`
	Chapter     ChapterViewMeta       `json:"chapter"`
	Content     string                `json:"content"`
	Placeholder *ChapterPlaceholder   `json:"placeholder,omitempty"`
	Navigation  ChapterViewNavigation `json:"navigation"`
	Comic       ChapterComicView      `json:"comic"`
}

type ChapterViewMeta struct {
	ID           uint   `json:"id"`
	NovelID      uint   `json:"novel_id"`
	ChapterNo    int    `json:"chapter_no"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	Status       string `json:"status"`
	RewriteCount int    `json:"rewrite_count"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type ChapterPlaceholder struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type ChapterNavItem struct {
	ChapterNo int    `json:"chapter_no"`
	Title     string `json:"title"`
	State     string `json:"state"`
}

type ChapterViewNavigation struct {
	Prev *ChapterNavItem `json:"prev"`
	Next *ChapterNavItem `json:"next"`
}

type ChapterComicView struct {
	PageCount int             `json:"page_count"`
	ImageURLs []string        `json:"image_urls"`
	Pages     []ViewComicPage `json:"pages"`
}

func (h *ChapterHandler) View(c *gin.Context) {
	novelID, _ := strconv.Atoi(c.Param("id"))
	chapterNo, chapterNoErr := strconv.Atoi(c.Param("no"))

	novel, err := h.nSvc.GetByID(uint(novelID))
	if err != nil {
		v1.NotFound(c, "作品不存在")
		return
	}

	var outline model.Outline
	outlineErr := h.db.Where("novel_id = ?", novelID).Order("version DESC").First(&outline).Error
	plan := []ChapterPlanSnapshot{}
	if outlineErr == nil {
		plan = parseChapterPlanSnapshots(outline.ChapterPlan)
	}

	planByNo := make(map[int]ChapterPlanSnapshot, len(plan))
	generatedByNo := make(map[int]model.Chapter)
	for _, item := range plan {
		planByNo[item.ChapterNo] = item
	}

	var generatedChapters []model.Chapter
	h.db.Where("novel_id = ?", novelID).Order("chapter_no").Find(&generatedChapters)
	for _, ch := range generatedChapters {
		generatedByNo[ch.ChapterNo] = ch
	}

	orderedSet := make(map[int]struct{}, len(plan)+len(generatedChapters))
	orderedNos := make([]int, 0, len(plan)+len(generatedChapters))
	for _, item := range plan {
		if _, exists := orderedSet[item.ChapterNo]; exists {
			continue
		}
		orderedSet[item.ChapterNo] = struct{}{}
		orderedNos = append(orderedNos, item.ChapterNo)
	}
	for _, ch := range generatedChapters {
		if _, exists := orderedSet[ch.ChapterNo]; exists {
			continue
		}
		orderedSet[ch.ChapterNo] = struct{}{}
		orderedNos = append(orderedNos, ch.ChapterNo)
	}
	sort.Ints(orderedNos)

	var chapter *model.Chapter
	var chapterErr error
	if chapterNoErr == nil && chapterNo > 0 {
		chapter, chapterErr = h.cSvc.GetByNovelAndNo(uint(novelID), chapterNo)
	}
	if chapterErr != nil && chapterErr != gorm.ErrRecordNotFound {
		v1.InternalServerError(c, chapterErr.Error())
		return
	}

	state := "invalid"
	chapterMeta := ChapterViewMeta{
		NovelID:   uint(novelID),
		ChapterNo: chapterNo,
		Status:    "invalid",
	}
	if chapterNoErr != nil {
		chapterMeta.ChapterNo = 0
	}
	content := ""
	var placeholder *ChapterPlaceholder
	comic := ChapterComicView{ImageURLs: []string{}, Pages: []ViewComicPage{}}

	if chapterNoErr != nil || chapterNo <= 0 {
		placeholder = &ChapterPlaceholder{
			Reason:  "invalid_chapter_no",
			Message: "章节号非法，必须为正整数。",
		}
	} else if chapterErr == nil {
		state = "generated"
		chapterMeta = ChapterViewMeta{
			ID:           chapter.ID,
			NovelID:      chapter.NovelID,
			ChapterNo:    chapter.ChapterNo,
			Title:        chapter.Title,
			Status:       chapter.Status,
			RewriteCount: chapter.RewriteCount,
			CreatedAt:    chapter.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    chapter.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if item, ok := planByNo[chapterNo]; ok {
			chapterMeta.Summary = item.Summary
			if chapterMeta.Title == "" {
				chapterMeta.Title = item.Title
			}
		}
		content = chapter.Content

		var dbPages []model.ComicPage
		h.db.Where("novel_id = ? AND chapter_id = ?", novelID, chapter.ID).Order("page_no").Find(&dbPages)
		allURLs := make([]string, 0)
		viewPages := make([]ViewComicPage, 0, len(dbPages))
		for _, p := range dbPages {
			urls := parseImageURLs(p.ImageURLs)
			allURLs = append(allURLs, urls...)
			viewPages = append(viewPages, ViewComicPage{
				PageNo:     p.PageNo,
				PanelCount: p.PanelCount,
				ImageURLs:  urls,
			})
		}
		comic = ChapterComicView{
			PageCount: len(dbPages),
			ImageURLs: allURLs,
			Pages:     viewPages,
		}
	} else if item, ok := planByNo[chapterNo]; ok {
		state = "planned"
		chapterMeta.Title = item.Title
		chapterMeta.Summary = item.Summary
		chapterMeta.Status = "planned"
		placeholder = &ChapterPlaceholder{
			Reason:  "not_generated",
			Message: "该章节已在大纲中规划，但正文尚未生成。",
		}
	} else {
		placeholder = &ChapterPlaceholder{
			Reason:  "chapter_not_found",
			Message: "章节号合法，但未命中已生成章节或大纲规划。",
		}
	}

	buildNav := func(targetNo int) *ChapterNavItem {
		if targetNo == 0 {
			return nil
		}
		if item, ok := planByNo[targetNo]; ok {
			state := "planned"
			if _, exists := generatedByNo[targetNo]; exists {
				state = "generated"
			}
			return &ChapterNavItem{ChapterNo: targetNo, Title: item.Title, State: state}
		}
		if ch, exists := generatedByNo[targetNo]; exists {
			return &ChapterNavItem{ChapterNo: targetNo, Title: ch.Title, State: "generated"}
		}
		return &ChapterNavItem{ChapterNo: targetNo, State: "invalid"}
	}

	prevNo, nextNo := 0, 0
	for idx, no := range orderedNos {
		if no != chapterNo {
			continue
		}
		if idx > 0 {
			prevNo = orderedNos[idx-1]
		}
		if idx+1 < len(orderedNos) {
			nextNo = orderedNos[idx+1]
		}
		break
	}

	v1.Success(c, ChapterViewResponse{
		Novel:       toNovelDTO(novel),
		State:       state,
		Chapter:     chapterMeta,
		Content:     content,
		Placeholder: placeholder,
		Navigation: ChapterViewNavigation{
			Prev: buildNav(prevNo),
			Next: buildNav(nextNo),
		},
		Comic: comic,
	})
}
