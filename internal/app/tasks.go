package app

import (
	"novelforge/internal/config"
	"novelforge/internal/handler"
	"novelforge/internal/service"
	"novelforge/internal/task"
)

func setupTasks(app *App, sseH *handler.SSEHandler) *task.Queue {
	provider := &configProvider{cfg: app.Config}
	taskHandler := service.NewTaskHandler(app.DB, app.Novel, app.Outline, app.Chapter, app.Comic, sseH, provider)

	factory := task.NewTaskFactory(taskHandler, sseH, provider)
	queue := factory.CreateQueue()
	taskHandler.SetTaskQueue(queue)

	app.Novel.SetTaskQueue(queue)
	app.Novel.SetMaxChapters(app.Config.MaxChapters)
	app.Novel.SetOutlineService(app.Outline)
	app.Novel.SetChapterService(app.Chapter)
	app.Novel.SetComicService(app.Comic)
	app.Novel.SetOnProgress(func(novelID uint, event task.Event) {
		sseH.PushEvent(novelID, event)
	})

	return queue
}

type configProvider struct {
	cfg *config.Config
}

func (p *configProvider) GetMaxChapters() int  { return p.cfg.MaxChapters }
func (p *configProvider) IsImageEnabled() bool { return p.cfg.ImageEnabled }
