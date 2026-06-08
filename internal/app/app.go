package app

import (
	"fmt"
	"novelforge/api/v1"
	"novelforge/internal/config"
	"novelforge/internal/handler"
	"novelforge/internal/model"
	"novelforge/internal/service"
	"novelforge/internal/task"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type App struct {
	DB        *gorm.DB
	Config    *config.Config
	Router    *gin.Engine
	TaskQueue *task.Queue

	Auth    *service.AuthService
	Outline *service.OutlineService
	Chapter *service.ChapterService
	Comic   *service.ComicService
	Novel   *service.NovelService
	Agent   *service.AgentService
}

func New() (*App, error) {
	cfg := config.Load()

	db, err := gorm.Open(sqlite.Open(cfg.DBPath+"?_journal_mode=WAL"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	db.AutoMigrate(
		&model.User{}, &model.AIConfig{}, &model.Novel{}, &model.Outline{},
		&model.Chapter{}, &model.ComicPage{}, &model.AgentTask{}, &model.Checkpoint{},
	)

	outlineSvc := service.NewOutlineService(db)
	chapterSvc := service.NewChapterService(db, outlineSvc)

	app := &App{
		DB:     db,
		Config: cfg,
		Auth:    service.NewAuthService(db),
		Outline: outlineSvc,
		Chapter: chapterSvc,
		Comic:   service.NewComicService(db, cfg.ImageDir),
		Novel:   service.NewNovelService(db),
		Agent:   service.NewAgentService(db),
	}

	sseH := handler.NewSSEHandler()
	app.TaskQueue = setupTasks(app, sseH)

	app.Router = v1.SetupRouter(cfg.SessionSecret, func(api *gin.RouterGroup) {
		registerRoutes(api, app, db, sseH)
	})
	app.Router.Static("/images", cfg.ImageDir)

	return app, nil
}

func (app *App) Run() error {
	fmt.Printf("NovelForge API running on http://localhost:%s\n", app.Config.Port)
	return app.Router.Run(":" + app.Config.Port)
}
