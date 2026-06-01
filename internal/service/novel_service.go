package service

import (
	"encoding/json"
	"novelforge/internal/ai"
	"novelforge/internal/model"
	"novelforge/internal/task"
	"gorm.io/gorm"
)

type NovelService struct {
	db         *gorm.DB
	outlineSvc *OutlineService
	chapterSvc *ChapterService
	comicSvc   *ComicService
	taskQ      *task.Queue
}

func NewNovelService(db *gorm.DB) *NovelService { return &NovelService{db: db} }

func (s *NovelService) SetOutlineService(svc *OutlineService) { s.outlineSvc = svc }
func (s *NovelService) SetChapterService(svc *ChapterService) { s.chapterSvc = svc }
func (s *NovelService) SetComicService(svc *ComicService)   { s.comicSvc = svc }
func (s *NovelService) SetTaskQueue(q *task.Queue)          { s.taskQ = q }

func (s *NovelService) List(userID uint) ([]model.Novel, error) {
	var novels []model.Novel
	err := s.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&novels).Error
	return novels, err
}

func (s *NovelService) GetByID(novelID uint) (*model.Novel, error) {
	var novel model.Novel
	err := s.db.First(&novel, novelID).Error
	return &novel, err
}

func (s *NovelService) Create(userID uint, title, summary, mode, imageMode string, aiConfigID uint) (*model.Novel, error) {
	novel := model.Novel{
		UserID:     userID,
		Title:      title,
		Summary:    summary,
		Mode:       mode,
		ImageMode:  imageMode,
		AIConfigID: aiConfigID,
	}
	if err := s.db.Create(&novel).Error; err != nil {
		return nil, err
	}
	return &novel, nil
}

func (s *NovelService) StartGeneration(novelID uint) error {
	novel, err := s.GetByID(novelID)
	if err != nil {
		return err
	}

	var outlineInput string
	switch novel.Mode {
	case "blindbox":
		titleResp, err := s.outlineSvc.provider.Chat(ai.ChatRequest{
			Messages: []ai.ChatMessage{
				{Role: "system", Content: "你是一个创意小说作者。请自动想一个引人入胜的小说题材和标题。输出纯 JSON：{\"title\":\"...\",\"summary\":\"一句话梗概\"}"},
			},
		})
		if err != nil {
			return err
		}
		var blindResult struct {
			Title   string `json:"title"`
			Summary string `json:"summary"`
		}
		json.Unmarshal([]byte(titleResp.Content), &blindResult)
		novel.Title = blindResult.Title
		novel.Summary = blindResult.Summary
		s.db.Save(novel)
		outlineInput = blindResult.Summary

	case "outline", "inspiration":
		outlineInput = novel.Summary

	default:
		outlineInput = novel.Summary
	}

	outline, err := s.outlineSvc.Generate(novel, outlineInput)
	if err != nil {
		return err
	}

	var chapterPlan []ChapterPlanItem
	json.Unmarshal([]byte(outline.ChapterPlan), &chapterPlan)

	for _, cp := range chapterPlan {
		s.taskQ.EnqueueWrite(task.Task{NovelID: novelID, ChapterNo: cp.ChapterNo})
	}

	novel.Status = "drafting"
	s.db.Save(novel)
	return nil
}

func CreateProviderFromConfig(cfg *model.AIConfig) ai.Provider {
	return ai.NewProvider(cfg.Provider, cfg.APIKey, cfg.BaseURL, cfg.TextModel, cfg.ImageModel)
}
