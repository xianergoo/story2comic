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
	json.Unmarshal([]byte(outline.ChapterPlan), &plan)
	chapter, err := h.chapterSvc.Write(t.NovelID, t.ChapterNo, plan, outline, t.Suggestion)
	if err != nil {
		fmt.Printf("ERROR write: novel=%d ch=%d err=%v\n", t.NovelID, t.ChapterNo, err)
		h.publisher.Push(t.NovelID, fmt.Sprintf(`{"type":"error","chapter_no":%d,"msg":"写文失败: %s"}`,
			t.ChapterNo, err.Error()))
		return err
	}
	h.publisher.Push(t.NovelID, fmt.Sprintf(`{"type":"chapter","chapter_no":%d,"status":"%s"}`,
		t.ChapterNo, chapter.Status))
	fmt.Printf("CHAPTER: novel=%d ch=%d status=%s\n", t.NovelID, t.ChapterNo, chapter.Status)
	h.db.Model(&model.Novel{}).Where("id = ?", t.NovelID).Update("updated_at", gorm.Expr("datetime('now')"))
	if chapter.Status == "done" && h.taskQueue != nil {
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

	err := h.comicSvc.Generate(chapter, novel, outline)
	if err == nil {
		h.publisher.Push(t.NovelID, fmt.Sprintf(`{"type":"comic","chapter_no":%d,"status":"done"}`, t.ChapterNo))
		fmt.Printf("COMIC: novel=%d ch=%d done\n", t.NovelID, t.ChapterNo)
	} else {
		fmt.Printf("ERROR comic: novel=%d ch=%d err=%v\n", t.NovelID, t.ChapterNo, err)
		h.publisher.Push(t.NovelID, fmt.Sprintf(`{"type":"error","chapter_no":%d,"msg":"生图失败: %s"}`,
			t.ChapterNo, err.Error()))
	}
	return err
}

// OnEnqueueImage 图片任务入队回调
func (h *TaskHandlerImpl) OnEnqueueImage(t task.Task) {
	fmt.Printf("图片任务入队: novel=%d chapter=%d\n", t.NovelID, t.ChapterNo)
}