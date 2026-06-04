package task

import "novelforge/internal/model"

// TaskHandler 任务处理器接口
type TaskHandler interface {
	// HandleWrite 处理写作任务
	HandleWrite(t Task) error
	
	// HandleImage 处理图片生成任务
	HandleImage(t Task) error
	
	// OnEnqueueImage 图片任务入队回调
	OnEnqueueImage(t Task)
}

// SSEPublisher SSE发布者接口
type SSEPublisher interface {
	Push(novelID uint, message string)
}

// ConfigProvider 配置提供者接口
type ConfigProvider interface {
	GetMaxChapters() int
	IsImageEnabled() bool
}

// DatabaseProvider 数据库提供者接口
type DatabaseProvider interface {
	GetDB() interface{} // 返回*gorm.DB或其他数据库连接
}

// ModelProvider 模型提供者接口
type ModelProvider interface {
	GetNovelByID(novelID uint) (*model.Novel, error)
	GetOutlineByNovel(novelID uint) (*model.Outline, error)
	GetChapterByNovelAndNo(novelID uint, chapterNo int) (*model.Chapter, error)
	GetAIConfigByID(configID uint) (*model.AIConfig, error)
	UpdateNovelStatus(novelID uint, updates map[string]interface{}) error
	CountChapters(novelID uint) (int64, error)
}

// AIProvider AI提供者接口
type AIProvider interface {
	SetProvider(provider interface{})
	WriteChapter(novelID uint, chapterNo int, plan interface{}, outline interface{}, suggestion string) (*model.Chapter, error)
	GenerateComic(chapter interface{}, novel interface{}, outline interface{}) error
}