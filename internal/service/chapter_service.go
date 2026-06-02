package service

import (
	"fmt"
	"strings"
	"novelforge/internal/ai"
	"novelforge/internal/model"

	"gorm.io/gorm"
)

const contextWindowSize = 3

type ChapterService struct {
	db         *gorm.DB
	outlineSvc *OutlineService
	provider   ai.Provider
}

func NewChapterService(db *gorm.DB, outlineSvc *OutlineService) *ChapterService {
	return &ChapterService{db: db, outlineSvc: outlineSvc}
}

func (s *ChapterService) SetProvider(p ai.Provider) { s.provider = p }

func (s *ChapterService) GetByNovelAndNo(novelID uint, chapterNo int) (*model.Chapter, error) {
	var ch model.Chapter
	err := s.db.Where("novel_id = ? AND chapter_no = ?", novelID, chapterNo).First(&ch).Error
	return &ch, err
}

func (s *ChapterService) Write(novelID uint, chapterNo int, chapterPlan []ChapterPlanItem, outline *model.Outline, suggestion string) (*model.Chapter, error) {
	plan := chapterPlan[chapterNo-1]

	var prevChapters []model.Chapter
	s.db.Where("novel_id = ? AND chapter_no < ? AND status = ?", novelID, chapterNo, "done").
		Order("chapter_no DESC").Limit(contextWindowSize).Find(&prevChapters)

	var summaries []string
	for i := len(prevChapters) - 1; i >= 0; i-- {
		ch := prevChapters[i]
		summary := ch.Content
		if len(summary) > 200 {
			summary = summary[:200]
		}
		summaries = append(summaries, summary)
	}
	contextSnapshot := strings.Join(summaries, "\n---\n")

	systemPrompt := `你是一个专业小说作家。根据大纲和上下文续写下一章正文。保持文风一致，人物性格一致。输出纯正文，不要带章节标题。`
	userPrompt := fmt.Sprintf(
		"故事大纲：%s\n\n前文摘要：%s\n\n人物设定：%s\n\n世界观：%s\n\n本章节标题：%s\n本章节大纲：%s\n请写出本章正文（约800字）：",
		outline.Content, contextSnapshot, outline.CharacterSheets, outline.WorldSetting, plan.Title, plan.Summary,
	)
	if suggestion != "" {
		userPrompt += "\n\n特别注意（用户建议）：" + suggestion
	}

	resp, err := s.provider.Chat(ai.ChatRequest{
		Messages: []ai.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	})
	if err != nil {
		return nil, err
	}

	chapter := &model.Chapter{
		NovelID:         novelID,
		ChapterNo:       chapterNo,
		Title:           plan.Title,
		Content:         resp.Content,
		ContextSnapshot: contextSnapshot,
		RewriteCount:    0,
	}

	// CoherenceCheck
	checkResult, err := s.outlineSvc.CheckCoherence(
		contextSnapshot,
		outline.CharacterSheets,
		outline.WorldSetting,
		resp.Content,
	)
	if err == nil && !checkResult.Pass && checkResult.Score < 6 {
		chapter.Status = "coherence_check"
		chapter.RewriteCount = 1
	} else {
		chapter.Status = "done"
	}

	s.db.Create(chapter)
	return chapter, nil
}
