package app

import (
	"net/http"

	"novelforge/api/v1"
	"novelforge/internal/handler"
	"novelforge/internal/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registerRoutes(api *gin.RouterGroup, app *App, db *gorm.DB, sseH *handler.SSEHandler) {
	api.GET("/health", func(c *gin.Context) {
		v1.Success(c, gin.H{"status": "ok", "version": "1.0.0"})
	})

	auth := api.Group("/auth")
	registerAuthRoutes(auth, app, db)

	authed := api.Group("")
	authed.Use(authMiddleware)
	registerBusinessRoutes(authed, app, db, sseH)
}

func registerAuthRoutes(auth *gin.RouterGroup, app *App, db *gorm.DB) {
	auth.POST("/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			v1.ValidationError(c, err.Error())
			return
		}
		user, err := app.Auth.Authenticate(req.Username, req.Password)
		if err != nil {
			v1.Unauthorized(c, "用户名或密码错误")
			return
		}
		session := sessions.Default(c)
		session.Set("user_id", user.ID)
		session.Set("username", user.Username)
		session.Save()
		v1.SuccessWithMessage(c, "登录成功", gin.H{"id": user.ID, "username": user.Username})
	})

	auth.POST("/register", func(c *gin.Context) {
		var req struct {
			Username        string `json:"username" binding:"required"`
			Password        string `json:"password" binding:"required"`
			ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			v1.ValidationError(c, err.Error())
			return
		}
		exists, _ := app.Auth.UserExists(req.Username)
		if exists {
			v1.Error(c, http.StatusConflict, "用户名已存在")
			return
		}
		user, err := app.Auth.CreateUser(req.Username, req.Password)
		if err != nil {
			v1.InternalServerError(c, "创建用户失败")
			return
		}
		session := sessions.Default(c)
		session.Set("user_id", user.ID)
		session.Set("username", user.Username)
		session.Save()
		v1.SuccessWithMessage(c, "注册成功", gin.H{"id": user.ID, "username": user.Username})
	})

	auth.POST("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		v1.Success(c, nil)
	})

	auth.GET("/me", func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			v1.Unauthorized(c, "未登录")
			return
		}
		var user model.User
		if err := db.First(&user, userID).Error; err != nil {
			v1.NotFound(c, "用户不存在")
			return
		}
		v1.Success(c, gin.H{"id": user.ID, "username": user.Username, "created_at": user.CreatedAt})
	})
}

func registerBusinessRoutes(r *gin.RouterGroup, app *App, db *gorm.DB, sseH *handler.SSEHandler) {
	novelH := handler.NewNovelHandler(app.Novel, db)
	chapterH := handler.NewChapterHandler(db, app.Novel, app.Chapter)
	aiConfigH := handler.NewAIConfigHandler(db)
	agentH := handler.NewAgentHandler(app.Agent, app.Novel)

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
