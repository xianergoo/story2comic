package main

import (
	"encoding/json"
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"novelforge/internal/config"
	"novelforge/internal/handler"
	"novelforge/internal/middleware"
	"novelforge/internal/model"
	"novelforge/internal/service"
	"novelforge/internal/task"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(sqlite.Open(cfg.DBPath+"?_journal_mode=WAL"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	db.AutoMigrate(
		&model.User{},
		&model.AIConfig{},
		&model.Novel{},
		&model.Outline{},
		&model.Chapter{},
		&model.ComicPage{},
	)

	authSvc := service.NewAuthService(db)
	outlineSvc := service.NewOutlineService(db)
	chapterSvc := service.NewChapterService(db, outlineSvc)
	comicSvc := service.NewComicService(db, cfg.ImageDir)
	novelSvc := service.NewNovelService(db)

	sseH := handler.NewSSEHandler()

	var taskQ *task.Queue
	taskQ = task.New(
		func(t task.Task) error {
			novel, _ := novelSvc.GetByID(t.NovelID)
			outline, _ := outlineSvc.GetByNovel(t.NovelID)

			var aiCfg model.AIConfig
			db.First(&aiCfg, novel.AIConfigID)
			p := service.CreateProviderFromConfig(&aiCfg)
			outlineSvc.SetProvider(p)
			chapterSvc.SetProvider(p)

			var plan []service.ChapterPlanItem
			json.Unmarshal([]byte(outline.ChapterPlan), &plan)
			chapter, err := chapterSvc.Write(t.NovelID, t.ChapterNo, plan, outline)
			if err != nil {
				return err
			}
			sseH.Push(t.NovelID, fmt.Sprintf(`{"type":"chapter","chapter_no":%d,"status":"done"}`, t.ChapterNo))
			if chapter.Status == "done" {
				taskQ.EnqueueImage(task.Task{NovelID: t.NovelID, ChapterNo: t.ChapterNo, Type: task.TaskImage})
			}
			return nil
		},
		func(t task.Task) error {
			novel, _ := novelSvc.GetByID(t.NovelID)
			outline, _ := outlineSvc.GetByNovel(t.NovelID)
			chapter, _ := chapterSvc.GetByNovelAndNo(t.NovelID, t.ChapterNo)

			var aiCfg model.AIConfig
			db.First(&aiCfg, novel.AIConfigID)
			p := service.CreateProviderFromConfig(&aiCfg)
			comicSvc.SetProvider(p)

			err := comicSvc.Generate(chapter, novel, outline)
			if err == nil {
				sseH.Push(t.NovelID, fmt.Sprintf(`{"type":"comic","chapter_no":%d,"status":"done"}`, t.ChapterNo))
			}
			return err
		},
	)

	novelSvc.SetTaskQueue(taskQ)
	novelSvc.SetOutlineService(outlineSvc)
	novelSvc.SetChapterService(chapterSvc)
	novelSvc.SetComicService(comicSvc)

	authH := handler.NewAuthHandler(authSvc)
	novelH := handler.NewNovelHandler(novelSvc, db)
	chapterH := handler.NewChapterHandler(db, novelSvc, chapterSvc)
	aiConfigH := handler.NewAIConfigHandler(db)

	r := gin.Default()
	r.SetTrustedProxies(nil)

	store := cookie.NewStore([]byte(cfg.SessionSecret))
	r.Use(sessions.Sessions("novelforge_session", store))

	r.LoadHTMLGlob("templates/*.html")

	r.GET("/login", middleware.GuestOnly(), authH.LoginPage)
	r.POST("/login", middleware.GuestOnly(), authH.Login)
	r.GET("/register", middleware.GuestOnly(), authH.RegisterPage)
	r.POST("/register", middleware.GuestOnly(), authH.Register)
	r.GET("/logout", authH.Logout)

	auth := r.Group("/", middleware.AuthRequired())
	{
		auth.GET("/", novelH.Home)
		auth.POST("/novel", novelH.Create)
		auth.GET("/novel/:id", novelH.Detail)
		auth.GET("/novel/:id/chapter/:no", chapterH.View)
		auth.GET("/api/sse", sseH.Subscribe)
		auth.GET("/ai-config", aiConfigH.Page)
		auth.POST("/ai-config", aiConfigH.Create)
		auth.DELETE("/ai-config/:id", aiConfigH.Delete)
	}

	r.Static("/images", cfg.ImageDir)
	r.Static("/static", "./static")

	fmt.Printf("NovelForge running on http://localhost:%s\n", cfg.Port)
	r.Run(":" + cfg.Port)
}
