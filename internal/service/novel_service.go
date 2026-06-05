package service

import (
	"encoding/json"
	"fmt"
	"novelforge/internal/ai"
	"novelforge/internal/model"
	"novelforge/internal/task"
	"gorm.io/gorm"
)

type ProgressCallback func(novelID uint, event task.Event)

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
	s.publishNovelStatus(novelID, novel.Status, novel.TextStatus, novel.ImageStatus, "generation_started")
	s.publishLog(novelID, "generation_start", "开始生成")

	var aiCfg model.AIConfig
	if err := s.db.First(&aiCfg, novel.AIConfigID).Error; err != nil {
		s.publishError(novelID, "load_ai_config", "AI 配置不存在，请先在 AI配置页面 中添加")
		return fmt.Errorf("AI 配置不存在，请先在 AI配置页面 中添加")
	}
	fmt.Printf("GEN-CFG: provider=%s model=%s\n", aiCfg.Provider, aiCfg.TextModel)
	provider := CreateProviderFromConfig(&aiCfg)
	s.outlineSvc.SetProvider(provider)
	s.chapterSvc.SetProvider(provider)

	var outlineInput string
	switch novel.Mode {
	case "blindbox":
		s.publishLog(novelID, "blindbox_title_start", "AI 正在构思标题和故事梗概...")
		titleResp, err := s.outlineSvc.provider.Chat(ai.ChatRequest{
			Messages: []ai.ChatMessage{
				{Role: "system", Content: "你是一个创意小说作者。请自动想一个引人入胜的小说题材和标题。输出纯 JSON：{\"title\":\"...\",\"summary\":\"一句话梗概\"}"},
			},
		})
		if err != nil {
			s.publishError(novelID, "blindbox_title_failed", "标题生成失败: "+err.Error())
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
		s.publishLog(novelID, "blindbox_title_done", fmt.Sprintf("标题: %s / %s", blindResult.Title, blindResult.Summary))

	case "outline", "inspiration":
		outlineInput = novel.Summary

	default:
		outlineInput = novel.Summary
	}

	s.publishNovelStatus(novelID, novel.Status, novel.TextStatus, novel.ImageStatus, "outline_started")
	s.publishLog(novelID, "outline_start", "正在生成故事大纲...")
	outline, err := s.outlineSvc.GenerateStream(novel, outlineInput, func(chunk string, done bool) {
		s.publishOutlineStream(novelID, chunk, done)
	})
	if err != nil {
		fmt.Printf("GEN-ERR: outline failed: %v\n", err)
		s.publishNovelStatus(novelID, novel.Status, novel.TextStatus, novel.ImageStatus, "outline_failed")
		s.publishError(novelID, "outline_failed", "大纲生成失败: "+err.Error())
		return err
	}
	fmt.Printf("GEN-OUTLINE: done\n")
	s.publishNovelStatus(novelID, novel.Status, novel.TextStatus, novel.ImageStatus, "outline_completed")
	s.publishLog(novelID, "outline_done", "大纲生成完成")

	var chapterPlan []ChapterPlanItem
	json.Unmarshal([]byte(outline.ChapterPlan), &chapterPlan)
	planned := len(chapterPlan)
	if s.maxChapters > 0 && planned > s.maxChapters {
		planned = s.maxChapters
	}
	s.publishLog(novelID, "chapter_queue_start", fmt.Sprintf("共 %d 章，开始逐章写作...", planned))

	queued := 0
	for i, cp := range chapterPlan {
		if s.maxChapters > 0 && i >= s.maxChapters {
			break
		}
		s.taskQ.EnqueueWrite(task.Task{NovelID: novelID, ChapterNo: cp.ChapterNo})
		queued++
	}

	if queued == 0 {
		novel.Status = "completed"
		novel.TextStatus = "done"
		s.db.Save(novel)
		s.publishNovelStatus(novelID, novel.Status, novel.TextStatus, novel.ImageStatus, "generation_completed")
		s.publishProgressSummaryForNovel(novelID)
		s.publishLog(novelID, "generation_completed", "没有可写章节，生成完成")
		return nil
	}

	novel.Status = "drafting"
	novel.TextStatus = "writing"
	s.db.Save(novel)
	s.publishNovelStatus(novelID, novel.Status, novel.TextStatus, novel.ImageStatus, "all_chapters_queued")
	s.publishProgressSummaryForNovel(novelID)
	s.publishLog(novelID, "all_chapters_queued", fmt.Sprintf("已入队 %d 个章节任务", queued))
	return nil
}

func (s *NovelService) publish(novelID uint, event task.Event) {
	if s.onProgress != nil {
		s.onProgress(novelID, event)
	}
}

func (s *NovelService) publishNovelStatus(novelID uint, status, textStatus, imageStatus, step string) {
	s.publish(novelID, task.NewEvent(task.EventTypeNovelStatus, map[string]any{
		"status":       status,
		"text_status":  textStatus,
		"image_status": imageStatus,
		"step":         step,
	}))
}

func (s *NovelService) publishProgressSummary(novelID uint, planned, textDone, queued, currentChapterNo int) {
	s.publish(novelID, task.NewEvent(task.EventTypeProgressSummary, map[string]any{
		"planned":            planned,
		"text_done":          textDone,
		"queued":             queued,
		"current_chapter_no": currentChapterNo,
	}))
}

func (s *NovelService) publishLog(novelID uint, step, message string) {
	s.publish(novelID, task.NewEvent(task.EventTypeLog, map[string]any{
		"step":    step,
		"message": message,
	}))
}

func (s *NovelService) publishOutlineStream(novelID uint, chunk string, done bool) {
	s.publish(novelID, task.NewEvent(task.EventTypeOutlineStream, map[string]any{
		"content": chunk,
		"done":    done,
	}))
}

func (s *NovelService) publishError(novelID uint, step, message string) {
	s.publish(novelID, task.NewEvent(task.EventTypeError, map[string]any{
		"step":    step,
		"message": message,
	}))
}

func (s *NovelService) EnqueueChapterWithSuggestion(novelID uint, chapterNo int, suggestion string) {
	novel, _ := s.GetByID(novelID)
	imageStatus := "idle"
	if novel != nil {
		imageStatus = novel.ImageStatus
	}

	s.db.Model(&model.Novel{}).Where("id = ?", novelID).Updates(map[string]any{
		"text_status": "writing", "status": "drafting",
	})
	s.publishNovelStatus(novelID, "drafting", "writing", imageStatus, "resume_requested")
	s.publishLog(novelID, "resume_requested", fmt.Sprintf("请求重写第 %d 章", chapterNo))
	s.taskQ.EnqueueWrite(task.Task{NovelID: novelID, ChapterNo: chapterNo, Suggestion: suggestion})
}

func (s *NovelService) MarkCompleted(novelID uint) {
	novel, _ := s.GetByID(novelID)
	imageStatus := "idle"
	if novel != nil {
		imageStatus = novel.ImageStatus
	}
	planned, textDone, _, _, err := s.loadProgressSummary(novelID)
	if err != nil {
		planned, textDone = 0, 0
	}
	if planned > 0 && textDone < planned {
		textDone = planned
	}

	s.db.Model(&model.Novel{}).Where("id = ?", novelID).Updates(map[string]any{"status": "completed", "text_status": "done"})
	s.publishNovelStatus(novelID, "completed", "done", imageStatus, "generation_completed")
	s.publishProgressSummary(novelID, planned, textDone, 0, 0)
	s.publishLog(novelID, "generation_completed", "作品生成完成")
}

func CreateProviderFromConfig(cfg *model.AIConfig) ai.Provider {
	return ai.NewProvider(cfg.Provider, cfg.APIKey, cfg.BaseURL, cfg.TextModel, cfg.ImageModel)
}

func (s *NovelService) Resume(novelID uint) error {
	novel, err := s.GetByID(novelID)
	if err != nil {
		return err
	}
	s.publishNovelStatus(novelID, novel.Status, novel.TextStatus, novel.ImageStatus, "resume_requested")
	s.publishLog(novelID, "resume_requested", "继续生成未完成章节")

	outline, err := s.outlineSvc.GetByNovel(novelID)
	if err != nil {
		s.publishError(novelID, "load_outline", "大纲不存在")
		return fmt.Errorf("大纲不存在")
	}
	var plan []ChapterPlanItem
	json.Unmarshal([]byte(outline.ChapterPlan), &plan)

	queued := 0
	for i, cp := range plan {
		if s.maxChapters > 0 && i >= s.maxChapters {
			break
		}
		var existing model.Chapter
		err := s.db.Where("novel_id = ? AND chapter_no = ? AND status = ?", novelID, cp.ChapterNo, "done").First(&existing).Error
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				s.publishError(novelID, "resume_check_chapter", "检查章节状态失败: "+err.Error())
				return err
			}
			s.taskQ.EnqueueWrite(task.Task{NovelID: novelID, ChapterNo: cp.ChapterNo})
			queued++
		}
	}
	if queued == 0 {
		s.db.Model(&model.Novel{}).Where("id = ?", novelID).Updates(map[string]any{
			"status":      "completed",
			"text_status": "done",
		})
		s.publishNovelStatus(novelID, "completed", "done", novel.ImageStatus, "generation_completed")
		s.publishProgressSummaryForNovel(novelID)
		s.publishLog(novelID, "generation_completed", "没有待恢复章节，生成已完成")
		return nil
	}
	s.db.Model(&model.Novel{}).Where("id = ?", novelID).Updates(map[string]any{
		"status":      "drafting",
		"text_status": "writing",
	})
	s.publishNovelStatus(novelID, "drafting", "writing", novel.ImageStatus, "all_chapters_queued")
	s.publishProgressSummaryForNovel(novelID)
	s.publishLog(novelID, "all_chapters_queued", fmt.Sprintf("恢复入队 %d 个章节任务", queued))
	return nil
}

func (s *NovelService) publishProgressSummaryForNovel(novelID uint) {
	planned, textDone, queued, currentChapterNo, err := s.loadProgressSummary(novelID)
	if err != nil {
		return
	}
	s.publishProgressSummary(novelID, planned, textDone, queued, currentChapterNo)
}

func (s *NovelService) loadProgressSummary(novelID uint) (planned, textDone, queued, currentChapterNo int, err error) {
	outline, err := s.outlineSvc.GetByNovel(novelID)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	var plan []ChapterPlanItem
	json.Unmarshal([]byte(outline.ChapterPlan), &plan)
	if s.maxChapters > 0 && len(plan) > s.maxChapters {
		plan = plan[:s.maxChapters]
	}
	planned = len(plan)
	if planned == 0 {
		return 0, 0, 0, 0, nil
	}

	var chapters []model.Chapter
	if err := s.db.Where("novel_id = ?", novelID).Find(&chapters).Error; err != nil {
		return 0, 0, 0, 0, err
	}
	doneChapters := make(map[int]bool, len(chapters))
	for _, chapter := range chapters {
		if chapter.Status == "done" {
			doneChapters[chapter.ChapterNo] = true
		}
	}

	for _, cp := range plan {
		if doneChapters[cp.ChapterNo] {
			textDone++
			continue
		}
		queued++
		if currentChapterNo == 0 {
			currentChapterNo = cp.ChapterNo
		}
	}

	return planned, textDone, queued, currentChapterNo, nil
}
