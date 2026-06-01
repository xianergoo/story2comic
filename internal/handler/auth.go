package handler

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"novelforge/internal/service"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc} }

func (h *AuthHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func (h *AuthHandler) RegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{})
}

func (h *AuthHandler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	user, err := h.svc.Login(username, password)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": err.Error()})
		return
	}
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Save()
	c.Redirect(http.StatusFound, "/")
}

func (h *AuthHandler) Register(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	confirm := c.PostForm("confirm_password")
	if password != confirm {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{"error": "两次密码不一致"})
		return
	}
	_, err := h.svc.Register(username, password)
	if err != nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, "/login")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}
