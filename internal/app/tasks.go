package app

import (
	"novelforge/internal/config"
	"novelforge/internal/handler"
	"novelforge/internal/service"
	"novelforge/internal/task"

	"gorm.io/gorm"
)

func setupTasks(db *gorm.DB, cfg *config.Config, services *Services, sseH *handler.SSEHandler) *task.Queue {
	provider := &configProvider{cfg: cfg}
	taskHandler := service.NewTaskHandler(db, services.Novel, services.Outline, services.Chapter, services.Comic, sseH, provider)

	factory := task.NewTaskFactory(taskHandler, sseH, provider)
	queue := factory.CreateQueue()
	taskHandler.SetTaskQueue(queue)

	services.Novel.SetOnProgress(func(novelID uint, event task.Event) {
		sseH.PushEvent(novelID, event)
	})
	services.Novel.SetTaskQueue(queue)

	return queue
}

type configProvider struct {
	cfg *config.Config
}

func (p *configProvider) GetMaxChapters() int  { return p.cfg.MaxChapters }
func (p *configProvider) IsImageEnabled() bool { return p.cfg.ImageEnabled }
