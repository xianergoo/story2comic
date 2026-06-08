package app

import (
	"novelforge/api/v1"
	"novelforge/internal/handler"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registerRoutes(api *gin.RouterGroup, services *Services, db *gorm.DB, sseH *handler.SSEHandler) {
	api.GET("/health", func(c *gin.Context) {
		v1.Success(c, gin.H{"status": "ok", "version": "1.0.0"})
	})

	auth := api.Group("/auth")
	registerAuthRoutes(auth, services, db)

	authed := api.Group("")
	authed.Use(authMiddleware)
	registerBusinessRoutes(authed, services, db, sseH)
}

func registerAuthRoutes(auth *gin.RouterGroup, services *Services, db *gorm.DB) {
	authH := handler.NewAuthAPIHandler(services.Auth, db)
	auth.POST("/login", authH.Login)
	auth.POST("/register", authH.Register)
	auth.POST("/logout", authH.Logout)
	auth.GET("/me", authH.Me)
}

func registerBusinessRoutes(r *gin.RouterGroup, services *Services, db *gorm.DB, sseH *handler.SSEHandler) {
	novelH := handler.NewNovelHandler(services.Novel, db)
	chapterH := handler.NewChapterHandler(db, services.Novel, services.Chapter)
	aiConfigH := handler.NewAIConfigHandler(db)
	agentH := handler.NewAgentHandler(services.Agent, services.Novel)

	r.GET("/novels", novelH.Home)
	r.POST("/novels", novelH.Create)
	r.GET("/novels/:id", novelH.Detail)
	r.POST("/novels/:id/stop", novelH.Stop)
	r.POST("/novels/:id/resume", novelH.Resume)

	r.GET("/novels/:id/chapters/:no", chapterH.View)
	r.POST("/novels/:id/chapters/:no/regenerate", novelH.RegenerateChapter)

	r.GET("/ai-configs", aiConfigH.Page)
	r.POST("/ai-configs", aiConfigH.Create)
	r.PUT("/ai-configs/:id", aiConfigH.Update)
	r.DELETE("/ai-configs/:id", aiConfigH.Delete)

	r.POST("/agent/tasks", agentH.CreateTask)
	r.GET("/agent/tasks", agentH.ListTasks)
	r.GET("/agent/tasks/:task_id", agentH.GetTask)
	r.POST("/agent/tasks/:task_id/cancel", agentH.CancelTask)
	r.GET("/agent/tasks/:task_id/checkpoints", agentH.GetPendingCheckpoints)
	r.PUT("/agent/checkpoints/:checkpoint_id", agentH.UpdateCheckpoint)

	r.GET("/sse", sseH.Subscribe)
}

func authMiddleware(c *gin.Context) {
	session := sessions.Default(c)
	uid := session.Get("user_id")
	if uid == nil {
		v1.Unauthorized(c, "未登录")
		c.Abort()
		return
	}
	// 注入到 Gin context，供 handler 通过 c.GetUint("user_id") 读取
	switch v := uid.(type) {
	case uint:
		c.Set("user_id", v)
	case int:
		c.Set("user_id", uint(v))
	case float64:
		c.Set("user_id", uint(v))
	default:
		c.Set("user_id", uid)
	}
	c.Next()
}
