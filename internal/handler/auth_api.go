package handler

import (
	"fmt"
	"net/http"

	v1 "novelforge/api/v1"
	"novelforge/internal/model"
	"novelforge/internal/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthAPIHandler struct {
	auth *service.AuthService
	db   *gorm.DB
}

func NewAuthAPIHandler(auth *service.AuthService, db *gorm.DB) *AuthAPIHandler {
	return &AuthAPIHandler{auth: auth, db: db}
}

func (h *AuthAPIHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.ValidationError(c, err.Error())
		return
	}
	user, err := h.auth.Authenticate(req.Username, req.Password)
	if err != nil {
		v1.Unauthorized(c, "用户名或密码错误")
		return
	}
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("username", user.Username)
	if err := saveSession(session); err != nil {
		v1.InternalServerError(c, "保存登录状态失败")
		return
	}
	v1.SuccessWithMessage(c, "登录成功", gin.H{"id": user.ID, "username": user.Username})
}

func (h *AuthAPIHandler) Register(c *gin.Context) {
	var req struct {
		Username        string `json:"username" binding:"required"`
		Password        string `json:"password" binding:"required"`
		ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		v1.ValidationError(c, err.Error())
		return
	}
	exists, _ := h.auth.UserExists(req.Username)
	if exists {
		v1.Error(c, http.StatusConflict, "用户名已存在")
		return
	}
	user, err := h.auth.CreateUser(req.Username, req.Password)
	if err != nil {
		v1.InternalServerError(c, "创建用户失败")
		return
	}
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("username", user.Username)
	if err := saveSession(session); err != nil {
		v1.InternalServerError(c, "保存登录状态失败")
		return
	}
	v1.SuccessWithMessage(c, "注册成功", gin.H{"id": user.ID, "username": user.Username})
}

func (h *AuthAPIHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := saveSession(session); err != nil {
		v1.InternalServerError(c, "清理登录状态失败")
		return
	}
	v1.Success(c, nil)
}

func (h *AuthAPIHandler) Me(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		v1.Unauthorized(c, "未登录")
		return
	}
	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		v1.NotFound(c, "用户不存在")
		return
	}
	v1.Success(c, gin.H{"id": user.ID, "username": user.Username, "created_at": user.CreatedAt})
}

func saveSession(session sessions.Session) error {
	if err := session.Save(); err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return nil
}
