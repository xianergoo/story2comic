package service

import (
	"encoding/json"
	"fmt"
	"novelforge/internal/ai"
	"novelforge/internal/model"
	"novelforge/internal/task"
	"gorm.io/gorm"
)

type ProgressCallback func(novelID uint, step, detail string)

type NovelService struct {
	db          *gorm.DB
	outlineSvc  *OutlineService
	chapterSvc  *ChapterService
	comicSvc    *ComicService
	taskQ       *task.Queue
	onProgress  ProgressCallback
	maxChapters int
}

func NewNovelService(db *gorm.DB) *NovelService { return &NovelService{db: db} }

func (s *NovelService) SetMaxChapters(n int) { s.maxChapters = n }

func (s *NovelService) SetOnProgress(fn ProgressCallback) { s.onProgress = fn }

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
	fmt.Printf("GEN-START: novel_id=%d\n", novelID)
	novel, err := s.GetByID(novelID)
	if err != nil {
		return err
	}
	s.push(novelID, "start", "开始生成")

	var aiCfg model.AIConfig
	if err := s.db.First(&aiCfg, novel.AIConfigID).Error; err != nil {
		return fmt.Errorf("AI 配置不存在，请先在 AI配置页面 中添加")
	}
	fmt.Printf("GEN-CFG: provider=%s model=%s\n", aiCfg.Provider, aiCfg.TextModel)
	provider := CreateProviderFromConfig(&aiCfg)
	s.outlineSvc.SetProvider(provider)
	s.chapterSvc.SetProvider(provider)

	var outlineInput string
	switch novel.Mode {
	case "blindbox":
		s.push(novelID, "title", "AI 正在构思标题和故事梗概...")
		titleResp, err := s.outlineSvc.provider.Chat(ai.ChatRequest{
			Messages: []ai.ChatMessage{
				{Role: "system", Content: "你是一个创意小说作者。请自动想一个引人入胜的小说题材和标题。输出纯 JSON：{\"title\":\"...\",\"summary\":\"一句话梗概\"}"},
			},
		})
		if err != nil {
			s.push(novelID, "error", "标题生成失败: "+err.Error())
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
		s.push(novelID, "title", fmt.Sprintf("标题: %s / %s", blindResult.Title, blindResult.Summary))

	case "outline", "inspiration":
		outlineInput = novel.Summary

	default:
		outlineInput = novel.Summary
	}

	s.push(novelID, "outline", "正在生成故事大纲...")
	outline, err := s.outlineSvc.Generate(novel, outlineInput)
	if err != nil {
		fmt.Printf("GEN-ERR: outline failed: %v\n", err)
		s.push(novelID, "error", "大纲生成失败: "+err.Error())
		return err
	}
	fmt.Printf("GEN-OUTLINE: done\n")
	s.push(novelID, "outline", "大纲生成完成")

	var chapterPlan []ChapterPlanItem
	json.Unmarshal([]byte(outline.ChapterPlan), &chapterPlan)
	s.push(novelID, "plan", fmt.Sprintf("共 %d 章，开始逐章写作...", len(chapterPlan)))

	for i, cp := range chapterPlan {
		if s.maxChapters > 0 && i >= s.maxChapters {
			break
		}
		s.taskQ.EnqueueWrite(task.Task{NovelID: novelID, ChapterNo: cp.ChapterNo})
	}

	novel.Status = "drafting"
	s.db.Save(novel)
	return nil
}

func (s *NovelService) push(novelID uint, step, detail string) {
	if s.onProgress != nil {
		s.onProgress(novelID, step, detail)
	}
}

func (s *NovelService) MarkCompleted(novelID uint) {
	s.db.Model(&model.Novel{}).Where("id = ?", novelID).Update("status", "completed")
}

func CreateProviderFromConfig(cfg *model.AIConfig) ai.Provider {
	return ai.NewProvider(cfg.Provider, cfg.APIKey, cfg.BaseURL, cfg.TextModel, cfg.ImageModel)
}

func (s *NovelService) Resume(novelID uint) error {
	outline, err := s.outlineSvc.GetByNovel(novelID)
	if err != nil {
		return fmt.Errorf("大纲不存在")
	}
	var plan []ChapterPlanItem
	json.Unmarshal([]byte(outline.ChapterPlan), &plan)

	for i, cp := range plan {
		if s.maxChapters > 0 && i >= s.maxChapters {
			break
		}
		var existing model.Chapter
		err := s.db.Where("novel_id = ? AND chapter_no = ? AND status = ?", novelID, cp.ChapterNo, "done").First(&existing).Error
		if err != nil {
			s.taskQ.EnqueueWrite(task.Task{NovelID: novelID, ChapterNo: cp.ChapterNo})
		}
	}
	s.db.Model(&model.Novel{}).Where("id = ?", novelID).Update("status", "drafting")
	return nil
}
