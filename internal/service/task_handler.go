package service

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"novelforge/internal/model"
	"novelforge/internal/task"
)

// TaskHandlerImpl 任务处理器实现
type TaskHandlerImpl struct {
	db         *gorm.DB
	novelSvc   *NovelService
	outlineSvc *OutlineService
	chapterSvc *ChapterService
	comicSvc   *ComicService
	publisher  task.SSEPublisher
	config     task.ConfigProvider
	taskQueue  *task.Queue
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(
	db *gorm.DB,
	novelSvc *NovelService,
	outlineSvc *OutlineService,
	chapterSvc *ChapterService,
	comicSvc *ComicService,
	publisher task.SSEPublisher,
	config task.ConfigProvider,
) *TaskHandlerImpl {
	return &TaskHandlerImpl{
		db:         db,
		novelSvc:   novelSvc,
		outlineSvc: outlineSvc,
		chapterSvc: chapterSvc,
		comicSvc:   comicSvc,
		publisher:  publisher,
		config:     config,
	}
}

// SetTaskQueue 设置任务队列引用
func (h *TaskHandlerImpl) SetTaskQueue(queue *task.Queue) {
	h.taskQueue = queue
}

func (h *TaskHandlerImpl) publishEvent(novelID uint, eventType task.EventType, payload map[string]any) {
	h.publisher.PushEvent(novelID, task.NewEvent(eventType, payload))
}

func (h *TaskHandlerImpl) publishChapterStatus(novelID uint, chapterNo int, status string) {
	h.publishEvent(novelID, task.EventTypeChapterStatus, map[string]any{
		"chapter_no": chapterNo,
		"status":     status,
	})
}

func (h *TaskHandlerImpl) publishChapterStream(novelID uint, chapterNo int, chunk string) {
	h.publishEvent(novelID, task.EventTypeChapterStream, map[string]any{
		"chapter_no": chapterNo,
		"content":    chunk,
	})
}

func (h *TaskHandlerImpl) publishComicStatus(novelID uint, chapterNo int, status string) {
	h.publishEvent(novelID, task.EventTypeComicStatus, map[string]any{
		"chapter_no": chapterNo,
		"status":     status,
	})
}

func (h *TaskHandlerImpl) publishLog(novelID uint, chapterNo int, step, message string) {
	payload := map[string]any{
		"step":    step,
		"message": message,
	}
	if chapterNo > 0 {
		payload["chapter_no"] = chapterNo
	}
	h.publishEvent(novelID, task.EventTypeLog, payload)
}

func (h *TaskHandlerImpl) publishError(novelID uint, chapterNo int, step, message string) {
	payload := map[string]any{
		"step":    step,
		"message": message,
	}
	if chapterNo > 0 {
		payload["chapter_no"] = chapterNo
	}
	h.publishEvent(novelID, task.EventTypeError, payload)
}

// HandleWrite 处理写作任务
func (h *TaskHandlerImpl) HandleWrite(t task.Task) error {
	// 检查作品状态：非 drafting 则跳过
	novel, _ := h.novelSvc.GetByID(t.NovelID)
	if novel.TextStatus != "writing" {
		fmt.Printf("SKIP: novel=%d text_status=%s\n", t.NovelID, novel.TextStatus)
		return nil
	}
	// 检查章节数上限
	var count int64
	h.db.Model(&model.Chapter{}).Where("novel_id = ?", t.NovelID).Count(&count)
	if int(count) >= h.config.GetMaxChapters() {
		fmt.Printf("SKIP: novel=%d chapters=%d >= max=%d\n", t.NovelID, count, h.config.GetMaxChapters())
		h.db.Model(&model.Novel{}).Where("id = ?", t.NovelID).Updates(map[string]any{
			"status": "completed", "text_status": "done",
		})
		h.publisher.Push(t.NovelID, fmt.Sprintf(`{"type":"progress","step":"done","detail":"已达到章节上限 %d，文字生成结束"}`, h.config.GetMaxChapters()))
		return nil
	}
	outline, _ := h.outlineSvc.GetByNovel(t.NovelID)

	var aiCfg model.AIConfig
	h.db.First(&aiCfg, novel.AIConfigID)
	p := CreateProviderFromConfig(&aiCfg)
	h.outlineSvc.SetProvider(p)
	h.chapterSvc.SetProvider(p)

	var plan []ChapterPlanItem
	if err := json.Unmarshal([]byte(outline.ChapterPlan), &plan); err != nil {
		fmt.Printf("ERROR write: novel=%d ch=%d parse_plan=%v\n", t.NovelID, t.ChapterNo, err)
		h.publishChapterStatus(t.NovelID, t.ChapterNo, "failed")
		h.publishError(t.NovelID, t.ChapterNo, "parse_chapter_plan", "章节规划解析失败: "+err.Error())
		return err
	}

	h.publishChapterStatus(t.NovelID, t.ChapterNo, "writing")
	h.publishLog(t.NovelID, t.ChapterNo, "chapter_writing_started", fmt.Sprintf("开始写作第 %d 章", t.ChapterNo))

	chapter, err := h.chapterSvc.WriteStream(t.NovelID, t.ChapterNo, plan, outline, t.Suggestion, func(chunk string) {
		h.publishChapterStream(t.NovelID, t.ChapterNo, chunk)
	})
	if err != nil {
		fmt.Printf("ERROR write: novel=%d ch=%d err=%v\n", t.NovelID, t.ChapterNo, err)
		h.publishChapterStatus(t.NovelID, t.ChapterNo, "failed")
		h.publishError(t.NovelID, t.ChapterNo, "chapter_writing_failed", "写文失败: "+err.Error())
		return err
	}
	h.publishChapterStatus(t.NovelID, t.ChapterNo, chapter.Status)
	h.publishLog(t.NovelID, t.ChapterNo, "chapter_writing_completed", fmt.Sprintf("第 %d 章写作完成", t.ChapterNo))
	fmt.Printf("CHAPTER: novel=%d ch=%d status=%s\n", t.NovelID, t.ChapterNo, chapter.Status)
	h.db.Model(&model.Novel{}).Where("id = ?", t.NovelID).Update("updated_at", gorm.Expr("datetime('now')"))
	if chapter.Status == "done" && h.taskQueue != nil {
		h.publishLog(t.NovelID, t.ChapterNo, "comic_queued", fmt.Sprintf("第 %d 章漫画任务已入队", t.ChapterNo))
		h.taskQueue.EnqueueImage(task.Task{NovelID: t.NovelID, ChapterNo: t.ChapterNo, Type: task.TaskImage})
	}
	return nil
}

// HandleImage 处理图片生成任务
func (h *TaskHandlerImpl) HandleImage(t task.Task) error {
	if !h.config.IsImageEnabled() {
		fmt.Printf("SKIP-IMAGE: disabled\n")
		return nil
	}
	novel, _ := h.novelSvc.GetByID(t.NovelID)
	if novel.ImageStatus != "generating" {
		fmt.Printf("SKIP-IMAGE: novel=%d image_status=%s\n", t.NovelID, novel.ImageStatus)
		return nil
	}
	outline, _ := h.outlineSvc.GetByNovel(t.NovelID)
	chapter, _ := h.chapterSvc.GetByNovelAndNo(t.NovelID, t.ChapterNo)

	var aiCfg model.AIConfig
	h.db.First(&aiCfg, novel.AIConfigID)
	p := CreateProviderFromConfig(&aiCfg)
	h.comicSvc.SetProvider(p)

	h.publishComicStatus(t.NovelID, t.ChapterNo, "generating")
	h.publishLog(t.NovelID, t.ChapterNo, "comic_generation_started", fmt.Sprintf("开始生成第 %d 章漫画", t.ChapterNo))
	err := h.comicSvc.Generate(chapter, novel, outline)
	if err == nil {
		h.publishComicStatus(t.NovelID, t.ChapterNo, "done")
		h.publishLog(t.NovelID, t.ChapterNo, "comic_generation_completed", fmt.Sprintf("第 %d 章漫画生成完成", t.ChapterNo))
		fmt.Printf("COMIC: novel=%d ch=%d done\n", t.NovelID, t.ChapterNo)
	} else {
		fmt.Printf("ERROR comic: novel=%d ch=%d err=%v\n", t.NovelID, t.ChapterNo, err)
		h.publishComicStatus(t.NovelID, t.ChapterNo, "failed")
		h.publishError(t.NovelID, t.ChapterNo, "comic_generation_failed", "生图失败: "+err.Error())
	}
	return err
}

// OnEnqueueImage 图片任务入队回调
func (h *TaskHandlerImpl) OnEnqueueImage(t task.Task) {
	fmt.Printf("图片任务入队: novel=%d chapter=%d\n", t.NovelID, t.ChapterNo)
}
