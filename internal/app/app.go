package app

import (
	"context"
	"fmt"
	"net/http"
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

type Services struct {
	Auth    *service.AuthService
	Outline *service.OutlineService
	Chapter *service.ChapterService
	Comic   *service.ComicService
	Novel   *service.NovelService
	Agent   *service.AgentService
}

type Runtime struct {
	SSEHandler *handler.SSEHandler
	TaskQueue  *task.Queue
}

type Application struct {
	DB       *gorm.DB
	Config   *config.Config
	Router   *gin.Engine
	Services *Services
	Runtime  *Runtime
	HTTP     *http.Server
}

func New(cfg *config.Config) (*Application, error) {
	db, err := openDB(cfg)
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}

	services := buildServices(db, cfg)
	runtime := buildRuntime(db, cfg, services)
	router := buildRouter(cfg, db, services, runtime)

	return &Application{
		DB:       db,
		Config:   cfg,
		Router:   router,
		Services: services,
		Runtime:  runtime,
		HTTP: &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: router,
		},
	}, nil
}

func openDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DBPath+"?_journal_mode=WAL"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	return db, nil
}

func migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{}, &model.AIConfig{}, &model.Novel{}, &model.Outline{},
		&model.Chapter{}, &model.ComicPage{}, &model.AgentTask{}, &model.Checkpoint{},
	); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}
	return nil
}

func buildServices(db *gorm.DB, cfg *config.Config) *Services {
	outlineSvc := service.NewOutlineService(db)
	chapterSvc := service.NewChapterService(db, outlineSvc)
	comicSvc := service.NewComicService(db, cfg.ImageDir)
	novelSvc := service.NewNovelService(service.NovelServiceDeps{
		DB:          db,
		Outline:     outlineSvc,
		Chapter:     chapterSvc,
		Comic:       comicSvc,
		MaxChapters: cfg.MaxChapters,
	})

	return &Services{
		Auth:    service.NewAuthService(db),
		Outline: outlineSvc,
		Chapter: chapterSvc,
		Comic:   comicSvc,
		Novel:   novelSvc,
		Agent:   service.NewAgentService(db),
	}
}

func buildRuntime(db *gorm.DB, cfg *config.Config, services *Services) *Runtime {
	sseHandler := handler.NewSSEHandler()
	taskQueue := setupTasks(db, cfg, services, sseHandler)

	return &Runtime{
		SSEHandler: sseHandler,
		TaskQueue:  taskQueue,
	}
}

func buildRouter(cfg *config.Config, db *gorm.DB, services *Services, runtime *Runtime) *gin.Engine {
	router := v1.SetupRouter(cfg.SessionSecret, func(api *gin.RouterGroup) {
		registerRoutes(api, services, db, runtime.SSEHandler)
	})
	router.Static("/images", cfg.ImageDir)
	return router
}

func (app *Application) Run() error {
	fmt.Printf("NovelForge API running on http://localhost:%s\n", app.Config.Port)
	err := app.HTTP.ListenAndServe()
	if err == nil || err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (app *Application) Shutdown(ctx context.Context) error {
	if app.Runtime != nil && app.Runtime.TaskQueue != nil {
		app.Runtime.TaskQueue.Close()
	}
	if app.HTTP == nil {
		return nil
	}
	err := app.HTTP.Shutdown(ctx)
	if err == nil || err == http.ErrServerClosed {
		return nil
	}
	return err
}
