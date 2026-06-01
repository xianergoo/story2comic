# NovelForge Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建一个 AI 驱动的自动小说写作 + 同步漫画生成 Web 应用，Go + Gin 单体架构。

**Architecture:** 单一 Gin 二进制服务，SQLite 存储，HTML 模板渲染（Tailwind + HTMX + Alpine.js CDN），Go channel 异步任务队列处理 AI 调用，SSE 实时推送进度。

**Tech Stack:** Go 1.22+, Gin, GORM + SQLite, Go html/template, Tailwind CSS CDN, HTMX, Alpine.js

**Spec:** `docs/superpowers/specs/2026-05-29-novelforge-design.md`

---

## 文件结构总览

```
auto_drama/
├── main.go
├── go.mod / go.sum
├── .env.example
├── data/                          # 运行时数据目录
├── internal/
│   ├── config/config.go           # 环境变量加载
│   ├── model/                     # GORM 模型
│   │   ├── user.go
│   │   ├── ai_config.go
│   │   ├── novel.go
│   │   ├── outline.go
│   │   ├── chapter.go
│   │   └── comic_page.go
│   ├── ai/                        # AI Provider 接口+实现
│   │   ├── provider.go
│   │   ├── openai.go
│   │   ├── qwen.go
│   │   └── custom.go
│   ├── task/queue.go              # 异步任务队列
│   ├── middleware/
│   │   └── auth.go                # Session + Auth 中间件
│   ├── service/                   # 业务逻辑
│   │   ├── auth_service.go
│   │   ├── novel_service.go
│   │   ├── outline_service.go
│   │   ├── chapter_service.go
│   │   ├── comic_service.go
│   │   └── coherence_check.go
│   └── handler/                   # HTTP 处理器
│       ├── auth.go
│       ├── ai_config.go
│       ├── novel.go
│       ├── chapter.go
│       ├── comic.go
│       └── sse.go
├── templates/
│   ├── layout.html
│   ├── login.html
│   ├── register.html
│   ├── home.html
│   ├── novel_detail.html
│   └── reader.html
└── static/                        # 不变静态文件
```

**职责分解：**
- `config/` — 读取 `.env`，提供全局 Config 结构体
- `model/` — 纯数据模型，GORM tag，无业务逻辑
- `ai/` — Provider 接口 + 三家实现，对外暴露 `Chat()` 和 `GenerateImage()`
- `task/` — 内存队列，双队列（writeQueue + imageQueue），各 1 worker
- `middleware/` — Session 验证 + Auth 注入
- `service/` — 核心业务：认证、小说流程协调、大纲/章节/漫画生成、一致性校验
- `handler/` — Gin handler，只做参数解析→调用 service→渲染模板/JSON

---

### Task 1: 项目脚手架

**Files:**
- Create: `go.mod`
- Create: `.env.example`
- Create: `data/.gitkeep`
- Create: `data/images/.gitkeep`
- Create: internal 目录及其子目录

- [ ] **Step 1: 初始化 Go module**

```bash
cd /Users/z3/workspace/explore/auto_drama && go mod init novelforge
```
Expected: 生成 go.mod

- [ ] **Step 2: 创建 .env.example**

```bash
# .env.example
PORT=8080
DB_PATH=data/novelforge.db
IMAGE_DIR=data/images
SESSION_SECRET=change-me-32-bytes-min
```
写入 `.env.example` 文件。

- [ ] **Step 3: 创建目录结构**

```bash
mkdir -p data/images \
  internal/config internal/model internal/ai internal/task \
  internal/middleware internal/service internal/handler \
  templates static
touch data/.gitkeep data/images/.gitkeep
```

- [ ] **Step 4: 拉取依赖**

```bash
cd /Users/z3/workspace/explore/auto_drama && \
  go get github.com/gin-gonic/gin \
         gorm.io/gorm \
         gorm.io/driver/sqlite \
         github.com/gin-contrib/sessions \
         github.com/gin-contrib/sessions/cookie \
         github.com/joho/godotenv \
         golang.org/x/crypto
```

- [ ] **Step 5: 验证编译**

```bash
echo 'package main; func main(){}' > /tmp/stub.go
cd /Users/z3/workspace/explore/auto_drama && go build -o /dev/null .
```
Expected: 成功（空 main.go 可编译）

---

### Task 2: 配置包

**Files:**
- Create: `internal/config/config.go`
- Modify: `main.go`

- [ ] **Step 1: 编写 Config 结构体与加载逻辑**

```go
package config

import (
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DBPath       string
	ImageDir     string
	SessionSecret string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:          getEnv("PORT", "8080"),
		DBPath:        getEnv("DB_PATH", "data/novelforge.db"),
		ImageDir:      getEnv("IMAGE_DIR", "data/images"),
		SessionSecret: getEnv("SESSION_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 2: 在 main.go 中验证**

```go
package main

import (
	"fmt"
	"novelforge/internal/config"
)

func main() {
	cfg := config.Load()
	fmt.Printf("port=%s db=%s\n", cfg.Port, cfg.DBPath)
}
```

```bash
go run main.go
```
Expected: 输出 `port=8080 db=data/novelforge.db`

---

### Task 3: 数据模型（User + AIConfig + Novel）

**Files:**
- Create: `internal/model/user.go`
- Create: `internal/model/ai_config.go`
- Create: `internal/model/novel.go`

- [ ] **Step 1: 编写 User 模型**

```go
package model

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Username     string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
```

- [ ] **Step 2: 编写 AIConfig 模型**

```go
package model

import "time"

type AIConfig struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"index;not null"`
	Name       string    `gorm:"not null"`
	Provider   string    `gorm:"not null"` // openai / qwen / custom
	APIKey     string    `gorm:"not null"`
	BaseURL    string    `gorm:"default:''"`
	TextModel  string    `gorm:"not null"`
	ImageModel string    `gorm:"not null"`
	IsDefault  bool      `gorm:"default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
```

- [ ] **Step 3: 编写 Novel 模型**

```go
package model

import "time"

type Novel struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"index;not null"`
	Title      string    `gorm:"not null"`
	Summary    string    `gorm:"default:''"`
	CoverURL   string    `gorm:"default:''"`
	Mode       string    `gorm:"not null"` // inspiration / outline / blindbox
	ImageMode  string    `gorm:"not null;default:single"` // single / multi
	Status     string    `gorm:"not null;default:drafting"` // drafting / completed / failed
	AIConfigID uint      `gorm:"default:null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
```

---

### Task 4: 数据模型（Outline + Chapter + ComicPage）

**Files:**
- Create: `internal/model/outline.go`
- Create: `internal/model/chapter.go`
- Create: `internal/model/comic_page.go`

- [ ] **Step 1: 编写 Outline 模型**

```go
package model

import "time"

type Outline struct {
	ID              uint      `gorm:"primaryKey"`
	NovelID         uint      `gorm:"index;not null"`
	Version         int       `gorm:"not null;default:1"`
	Content         string    `gorm:"not null"`
	CharacterSheets string    `gorm:"default:'{}'"` // JSON
	WorldSetting    string    `gorm:"default:'{}'"` // JSON
	ChapterPlan     string    `gorm:"default:'[]'"` // JSON
	CreatedAt       time.Time
}
```

- [ ] **Step 2: 编写 Chapter 模型**

```go
package model

import "time"

type Chapter struct {
	ID              uint      `gorm:"primaryKey"`
	NovelID         uint      `gorm:"index;not null"`
	ChapterNo       int       `gorm:"not null"`
	Title           string    `gorm:"not null"`
	Content         string    `gorm:"default:''"`
	Status          string    `gorm:"not null;default:pending"` // pending/writing/coherence_check/done/failed
	RewriteCount    int       `gorm:"default:0"`
	ContextSnapshot string    `gorm:"default:''"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
```

- [ ] **Step 3: 编写 ComicPage 模型**

```go
package model

import "time"

type ComicPage struct {
	ID              uint      `gorm:"primaryKey"`
	ChapterID       uint      `gorm:"index;not null"`
	NovelID         uint      `gorm:"index;not null"`
	PageNo          int       `gorm:"not null"`
	PanelCount      int       `gorm:"not null;default:4"`
	Script          string    `gorm:"default:'{}'"` // JSON: 分镜脚本
	ImageURLs       string    `gorm:"default:'[]'"` // JSON: 图片路径列表
	StyleDesc       string    `gorm:"default:''"`
	Status          string    `gorm:"not null;default:pending"` // pending/generating/check/done/failed
	RetryCount      int       `gorm:"default:0"`
	ContextSnapshot string    `gorm:"default:''"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
```

---

### Task 5: 数据库初始化 + 自动迁移

**Files:**
- Modify: `main.go`

- [ ] **Step 1: 编写数据库初始化函数**

在 `main.go` 中实现：

```go
import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"novelforge/internal/model"
)

func initDB(dbPath string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL"), &gorm.Config{})
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
	return db
}
```

- [ ] **Step 2: 在 main 中调用**

```go
func main() {
	cfg := config.Load()
	db := initDB(cfg.DBPath)
	fmt.Printf("DB migrated, %s\n", cfg.Port)
}
```

- [ ] **Step 3: 验证**

```bash
go run main.go && ls -la data/novelforge.db
```
Expected: 生成 `data/novelforge.db` 文件。

---

### Task 6: AI Provider 接口

**Files:**
- Create: `internal/ai/provider.go`

- [ ] **Step 1: 定义接口**

```go
package ai

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatResponse struct {
	Content string `json:"content"`
}

type ImageRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Size   string `json:"size"`
	N      int    `json:"n"`
}

type ImageResponse struct {
	URLs []string `json:"urls"` // 返回图片的临时 URL 或 base64
}

type Provider interface {
	Chat(req ChatRequest) (*ChatResponse, error)
	GenerateImage(req ImageRequest) (*ImageResponse, error)
}

func NewProvider(providerType, apiKey, baseURL, textModel, imageModel string) Provider {
	// ... will be implemented in subsequent tasks
	return nil
}
```

---

### Task 7: OpenAI Provider 实现

**Files:**
- Create: `internal/ai/openai.go`

- [ ] **Step 1: 实现 OpenAI Provider**

```go
package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	textModel  string
	imageModel string
	client     *http.Client
}

func newOpenAI(apiKey, baseURL, textModel, imageModel string) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIProvider{
		apiKey:     apiKey,
		baseURL:    baseURL,
		textModel:  textModel,
		imageModel: imageModel,
		client:     &http.Client{},
	}
}

func (p *OpenAIProvider) Chat(req ChatRequest) (*ChatResponse, error) {
	body := map[string]interface{}{
		"model":    p.textModel,
		"messages": req.Messages,
	}
	b, _ := json.Marshal(body)

	httpReq, _ := http.NewRequest("POST", p.baseURL+"/chat/completions", bytes.NewReader(b))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("openai chat error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(respBody, &result)

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}
	return &ChatResponse{Content: result.Choices[0].Message.Content}, nil
}

func (p *OpenAIProvider) GenerateImage(req ImageRequest) (*ImageResponse, error) {
	body := map[string]interface{}{
		"model":  p.imageModel,
		"prompt": req.Prompt,
		"size":   req.Size,
		"n":      req.N,
	}
	b, _ := json.Marshal(body)

	httpReq, _ := http.NewRequest("POST", p.baseURL+"/images/generations", bytes.NewReader(b))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("openai image error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	json.Unmarshal(respBody, &result)

	var urls []string
	for _, d := range result.Data {
		urls = append(urls, d.URL)
	}
	return &ImageResponse{URLs: urls}, nil
}
```

---

### Task 8: Qwen + Custom Provider 实现

**Files:**
- Create: `internal/ai/qwen.go`
- Create: `internal/ai/custom.go`
- Modify: `internal/ai/provider.go`（实现 NewProvider 工厂函数）

- [ ] **Step 1: 实现 Qwen Provider**

```go
package ai

type QwenProvider struct{ OpenAIProvider }
// Qwen 兼容 OpenAI API 格式，直接嵌入 OpenAIProvider

func newQwen(apiKey, baseURL, textModel, imageModel string) *QwenProvider {
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	return &QwenProvider{OpenAIProvider{
		apiKey:     apiKey,
		baseURL:    baseURL,
		textModel:  textModel,
		imageModel: imageModel,
		client:     &http.Client{},
	}}
}
// Chat 和 GenerateImage 继承自 OpenAIProvider
```

- [ ] **Step 2: 实现 Custom Provider**

```go
package ai

func newCustom(apiKey, baseURL, textModel, imageModel string) *OpenAIProvider {
	return newOpenAI(apiKey, baseURL, textModel, imageModel)
}
// Custom 完全兼容 OpenAI 格式，直接返回 OpenAIProvider
```

- [ ] **Step 3: 实现 NewProvider 工厂函数**

修改 `internal/ai/provider.go`：

```go
func NewProvider(providerType, apiKey, baseURL, textModel, imageModel string) Provider {
	switch providerType {
	case "openai":
		return newOpenAI(apiKey, baseURL, textModel, imageModel)
	case "qwen":
		return newQwen(apiKey, baseURL, textModel, imageModel)
	case "custom":
		return newCustom(apiKey, baseURL, textModel, imageModel)
	default:
		return newOpenAI(apiKey, baseURL, textModel, imageModel)
	}
}
```

---

### Task 9: 异步任务队列

**Files:**
- Create: `internal/task/queue.go`

- [ ] **Step 1: 实现双队列任务系统**

```go
package task

import "fmt"

type TaskType string

const (
	TaskWrite TaskType = "write" // 写文
	TaskImage TaskType = "image" // 生图
)

type Task struct {
	NovelID   uint
	ChapterNo int
	Type      TaskType
}

type Queue struct {
	writeChan chan Task
	imageChan chan Task
	writeFunc func(Task) error
	imageFunc func(Task) error
}

func New(writeFn, imageFn func(Task) error) *Queue {
	q := &Queue{
		writeChan: make(chan Task, 10),
		imageChan: make(chan Task, 10),
		writeFunc: writeFn,
		imageFunc: imageFn,
	}
	go q.writeWorker()
	go q.imageWorker()
	return q
}

func (q *Queue) EnqueueWrite(t Task) { q.writeChan <- t }
func (q *Queue) EnqueueImage(t Task) { q.imageChan <- t }

func (q *Queue) writeWorker() {
	for t := range q.writeChan {
		if err := q.writeFunc(t); err != nil {
			fmt.Printf("write task failed: novel=%d ch=%d err=%v\n", t.NovelID, t.ChapterNo, err)
		}
	}
}

func (q *Queue) imageWorker() {
	for t := range q.imageChan {
		if err := q.imageFunc(t); err != nil {
			fmt.Printf("image task failed: novel=%d ch=%d err=%v\n", t.NovelID, t.ChapterNo, err)
		}
	}
}
```

---

### Task 10: Session + Auth 中间件

**Files:**
- Create: `internal/middleware/auth.go`

- [ ] **Step 1: 实现中间件**

```go
package middleware

import (
	"net/http"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Set("user_id", userID.(uint))
		c.Next()
	}
}

func GuestOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if userID := session.Get("user_id"); userID != nil {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}
```

---

### Task 11: Auth Service

**Files:**
- Create: `internal/service/auth_service.go`

- [ ] **Step 1: 实现注册/登录/登出**

```go
package service

import (
	"errors"
	"novelforge/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct{ db *gorm.DB }

func NewAuthService(db *gorm.DB) *AuthService { return &AuthService{db} }

func (s *AuthService) Register(username, password string) (*model.User, error) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := model.User{Username: username, PasswordHash: string(hash)}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, errors.New("用户名已存在")
	}
	return &user, nil
}

func (s *AuthService) Login(username, password string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	return &user, nil
}
```

---

### Task 12: Auth Handler + HTML 模板

**Files:**
- Create: `internal/handler/auth.go`
- Create: `templates/layout.html`
- Create: `templates/login.html`
- Create: `templates/register.html`

- [ ] **Step 1: 编写 Auth Handler**

```go
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
```

- [ ] **Step 2: 编写 layout.html 基础布局**

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{block "title" .}}NovelForge{{end}}</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <script src="https://unpkg.com/htmx.org@1.9.10"></script>
  <script defer src="https://unpkg.com/alpinejs@3.13.5"></script>
</head>
<body class="bg-gray-50 min-h-screen">
  <nav class="bg-white shadow-sm border-b px-6 py-3 flex items-center justify-between">
    <a href="/" class="text-xl font-bold text-indigo-600">NovelForge</a>
    <div class="flex items-center gap-4">
      <a href="/" class="text-sm text-gray-600 hover:text-indigo-600">书架</a>
      <a href="/ai-config" class="text-sm text-gray-600 hover:text-indigo-600">AI配置</a>
      <a href="/logout" class="text-sm text-gray-400 hover:text-red-500">退出</a>
    </div>
  </nav>
  <main class="max-w-6xl mx-auto px-4 py-6">
    {{template "content" .}}
  </main>
</body>
</html>
```

- [ ] **Step 3: 编写 login.html**

```html
{{define "title"}}登录 - NovelForge{{end}}
{{define "content"}}
<div class="max-w-sm mx-auto mt-20">
  <h1 class="text-2xl font-bold mb-6 text-center">登录</h1>
  {{if .error}}<div class="bg-red-50 text-red-600 p-3 rounded mb-4 text-sm">{{.error}}</div>{{end}}
  <form hx-post="/login" hx-target="body" hx-push-url="false" class="space-y-4">
    <input name="username" placeholder="用户名" required class="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-400 outline-none">
    <input name="password" type="password" placeholder="密码" required class="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-400 outline-none">
    <button type="submit" class="w-full bg-indigo-600 text-white py-2 rounded-lg hover:bg-indigo-700 font-medium">登录</button>
  </form>
  <p class="text-center text-sm text-gray-500 mt-4">没有账号？<a href="/register" class="text-indigo-600">注册</a></p>
</div>
{{end}}
```

- [ ] **Step 4: 编写 register.html**（类似 login，含 confirm_password 字段，表单指向 `/register`）

---

### Task 13: AI Config Handler

**Files:**
- Create: `internal/handler/ai_config.go`
- Create: `templates/ai_config.html`（直接内联在 handler 文件中也可，尽量减少模板文件数；决定还是单独模板）

实际实现：AI Config 页面使用内联 HTML 嵌入 handler 渲染，减少模板文件。

- [ ] **Step 1: 实现 AI 配置 CRUD handler**

```go
package handler

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"novelforge/internal/model"
)

type AIConfigHandler struct{ db *gorm.DB }

func NewAIConfigHandler(db *gorm.DB) *AIConfigHandler { return &AIConfigHandler{db} }

func (h *AIConfigHandler) Page(c *gin.Context) {
	userID := c.GetUint("user_id")
	var configs []model.AIConfig
	h.db.Where("user_id = ?", userID).Find(&configs)
	c.HTML(http.StatusOK, "ai_config.html", gin.H{"configs": configs})
}

func (h *AIConfigHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")
	isDefault := c.PostForm("is_default") == "on"
	if isDefault {
		h.db.Model(&model.AIConfig{}).Where("user_id = ?", userID).Update("is_default", false)
	}
	cfg := model.AIConfig{
		UserID:     userID,
		Name:       c.PostForm("name"),
		Provider:   c.PostForm("provider"),
		APIKey:     c.PostForm("api_key"),
		BaseURL:    c.PostForm("base_url"),
		TextModel:  c.PostForm("text_model"),
		ImageModel: c.PostForm("image_model"),
		IsDefault:  isDefault,
	}
	h.db.Create(&cfg)
	c.Redirect(http.StatusFound, "/ai-config")
}

func (h *AIConfigHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID := c.GetUint("user_id")
	h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.AIConfig{})
	c.Redirect(http.StatusFound, "/ai-config")
}
```

- [ ] **Step 2: 编写 ai_config.html 模板**

直接渲染 AI 配置列表 + 新增表单（含 provider 下拉、api_key、base_url、text_model、image_model 字段）

---

### Task 14: Novel Service + CRUD Handler

**Files:**
- Create: `internal/service/novel_service.go`
- Create: `internal/handler/novel.go`

- [ ] **Step 1: 编写 NovelService（基础 CRUD）**

```go
package service

import (
	"novelforge/internal/model"
	"gorm.io/gorm"
)

type NovelService struct {
	db *gorm.DB
	// queue 接口在后续注入
}

func NewNovelService(db *gorm.DB) *NovelService { return &NovelService{db} }

func (s *NovelService) List(userID uint) ([]model.Novel, error) {
	var novels []model.Novel
	err := s.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&novels).Error
	return novels, err
}

func (s *NovelService) Get(id, userID uint) (*model.Novel, error) {
	var novel model.Novel
	err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&novel).Error
	return &novel, err
}

func (s *NovelService) Create(userID uint, title, summary, mode, imageMode string, aiConfigID uint) (*model.Novel, error) {
	novel := model.Novel{
		UserID:    userID,
		Title:     title,
		Summary:   summary,
		Mode:      mode,
		ImageMode: imageMode,
		AIConfigID: aiConfigID,
	}
	if err := s.db.Create(&novel).Error; err != nil {
		return nil, err
	}
	return &novel, nil
}
```

- [ ] **Step 2: 编写 Novel Handler**

```go
package handler

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"novelforge/internal/service"
)

type NovelHandler struct{ svc *service.NovelService }

func NewNovelHandler(svc *service.NovelService) *NovelHandler { return &NovelHandler{svc} }

func (h *NovelHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")
	title := c.PostForm("title")
	mode := c.PostForm("mode")
	imageMode := c.PostForm("image_mode")
	summary := c.PostForm("summary")
	aiConfigID, _ := strconv.Atoi(c.PostForm("ai_config_id"))
	h.svc.Create(userID, title, summary, mode, imageMode, uint(aiConfigID))
	c.Redirect(http.StatusFound, "/")
}
```

---

### Task 15: 首页 + 作品详情模板

**Files:**
- Create: `templates/home.html`
- Create: `templates/novel_detail.html`

- [ ] **Step 1: 编写 home.html（书架首页）**

展示用户所有作品卡片网格，每卡显示封面占位、标题、模式标签、进度。顶部「新建作品」按钮触发 Alpine.js 弹窗。弹窗内三种模式 Tab + 表单。

- [ ] **Step 2: 编写 novel_detail.html（作品详情页）**

左侧大纲树（可折叠章节列表），右侧漫画缩略图网格，顶部进度条。

---

### Task 16: Outline Service

**Files:**
- Create: `internal/service/outline_service.go`

- [ ] **Step 1: 实现大纲生成**

```go
package service

import (
	"encoding/json"
	"fmt"
	"novelforge/internal/ai"
	"novelforge/internal/model"
	"gorm.io/gorm"
)

type OutlineService struct {
	db       *gorm.DB
	provider ai.Provider
}

func NewOutlineService(db *gorm.DB) *OutlineService { return &OutlineService{db: db} }

func (s *OutlineService) SetProvider(p ai.Provider) { s.provider = p }

func (s *OutlineService) GetByNovel(novelID uint) (*model.Outline, error) {
	var o model.Outline
	err := s.db.Where("novel_id = ?", novelID).Order("version DESC").First(&o).Error
	return &o, err
}

type OutlineResult struct {
	Content         string          `json:"content"`
	CharacterSheets CharacterSheets `json:"character_sheets"`
	WorldSetting    string          `json:"world_setting"`
	ChapterPlan     []ChapterPlanItem `json:"chapter_plan"`
}

type CharacterSheets map[string]CharacterSheet
type CharacterSheet struct {
	Name       string `json:"name"`
	Role       string `json:"role"`
	Personality string `json:"personality"`
	Appearance string `json:"appearance"`
}
type ChapterPlanItem struct {
	ChapterNo int    `json:"chapter_no"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
}

func (s *OutlineService) Generate(novel *model.Novel, userInput string) (*model.Outline, error) {
	systemPrompt := `你是一个专业的小说策划，请根据用户输入生成完整的故事大纲。

输出 JSON 格式：
{
  "content": "故事大纲全文（500字左右）",
  "character_sheets": {"角色名": {"name":"...","role":"...","personality":"...","appearance":"..."}},
  "world_setting": "世界观描述",
  "chapter_plan": [{"chapter_no": 1, "title": "...", "summary": "..."}]
}`

	userPrompt := fmt.Sprintf("用户输入：%s", userInput)
	resp, err := s.provider.Chat(ai.ChatRequest{
		Messages: []ai.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	})
	if err != nil {
		return nil, err
	}

	var result OutlineResult
	json.Unmarshal([]byte(resp.Content), &result)

	csJSON, _ := json.Marshal(result.CharacterSheets)
	cpJSON, _ := json.Marshal(result.ChapterPlan)

	outline := &model.Outline{
		NovelID:         novel.ID,
		Content:         result.Content,
		CharacterSheets: string(csJSON),
		WorldSetting:    result.WorldSetting,
		ChapterPlan:     string(cpJSON),
	}
	s.db.Create(outline)
	return outline, nil
}
```

---

### Task 17: CoherenceCheck 校验服务

**Files:**
- Create: `internal/service/coherence_check.go`

- [ ] **Step 1: 实现一致性校验**

```go
package service

import (
	"encoding/json"
	"novelforge/internal/ai"
)

type CheckResult struct {
	Pass       bool     `json:"pass"`
	Score      int      `json:"score"`
	Issues     []string `json:"issues"`
	Suggestion string   `json:"suggestion"`
}

type ComicCheckResult struct {
	Pass        bool     `json:"pass"`
	Issues      []string `json:"issues"`
	RetryPrompt string   `json:"retry_prompt"`
}

func (s *OutlineService) CheckCoherence(
	previousSummaries string,
	characterSheets string,
	worldSetting string,
	chapterContent string,
) (*CheckResult, error) {
	systemPrompt := `你是一个专业的小说编辑，审查章节剧情连贯性。
检查维度：人物一致性、情节逻辑、世界观冲突。
输出 JSON：{"pass":bool,"score":0-10,"issues":["..."],"suggestion":"..."}`

	prompt := "前文摘要：" + previousSummaries +
		"\n人物设定：" + characterSheets +
		"\n世界观：" + worldSetting +
		"\n当前章节：" + chapterContent

	resp, err := s.provider.Chat(ai.ChatRequest{
		Messages: []ai.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}

	var result CheckResult
	json.Unmarshal([]byte(resp.Content), &result)
	return &result, nil
}
```

注意：`CheckCoherence` 方法属于 `OutlineService`，因为需要用到 provider。后续也可以独立为一个服务类，但为了减少文件数，先放 `OutlineService` 中。

---

### Task 18: Chapter Service（含上下文窗口管理）

**Files:**
- Create: `internal/service/chapter_service.go`

- [ ] **Step 1: 实现章节写作 + 上下文窗口**

```go
package service

import (
	"fmt"
	"strings"
	"novelforge/internal/ai"
	"novelforge/internal/model"
	"gorm.io/gorm"
)

const contextWindowSize = 3 // 保留前 3 章摘要

type ChapterService struct {
	db         *gorm.DB
	outlineSvc *OutlineService
	provider   ai.Provider
}

func NewChapterService(db *gorm.DB, outlineSvc *OutlineService) *ChapterService {
	return &ChapterService{db: db, outlineSvc: outlineSvc}
}

func (s *ChapterService) SetProvider(p ai.Provider) { s.provider = p }

func (s *ChapterService) GetByNovelAndNo(novelID uint, chapterNo int) (*model.Chapter, error) {
	var ch model.Chapter
	err := s.db.Where("novel_id = ? AND chapter_no = ?", novelID, chapterNo).First(&ch).Error
	return &ch, err
}

func (s *ChapterService) Write(novelID uint, chapterNo int, chapterPlan []ChapterPlanItem, outline *model.Outline) (*model.Chapter, error) {
	plan := chapterPlan[chapterNo-1]

	// 构建上下文：前 3 章摘要
	var prevChapters []model.Chapter
	s.db.Where("novel_id = ? AND chapter_no < ? AND status = ?", novelID, chapterNo, "done").
		Order("chapter_no DESC").Limit(contextWindowSize).Find(&prevChapters)

	var summaries []string
	for i := len(prevChapters) - 1; i >= 0; i-- {
		ch := prevChapters[i]
		summary := ch.Content
		if len(summary) > 200 {
			summary = summary[:200]
		}
		summaries = append(summaries, summary)
	}
	contextSnapshot := strings.Join(summaries, "\n---\n")

	systemPrompt := `你是一个专业小说作家。根据大纲和上下文续写下一章正文。保持文风一致，人物性格一致。`
	userPrompt := fmt.Sprintf("故事大纲：%s\n\n前文摘要：%s\n\n人物设定：%s\n\n世界观：%s\n\n本章节标题：%s\n本章节大纲：%s\n请写出本章正文（约800字）：",
		outline.Content, contextSnapshot, outline.CharacterSheets, outline.WorldSetting, plan.Title, plan.Summary)

	resp, err := s.provider.Chat(ai.ChatRequest{
		Messages: []ai.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	})
	if err != nil {
		return nil, err
	}

	// CoherenceCheck
	checkResult, err := s.outlineSvc.CheckCoherence(
		contextSnapshot,
		outline.CharacterSheets,
		outline.WorldSetting,
		resp.Content,
	)
	if err != nil {
		return nil, err
	}

	chapter := &model.Chapter{
		NovelID:         novelID,
		ChapterNo:       chapterNo,
		Title:           plan.Title,
		Content:         resp.Content,
		ContextSnapshot: contextSnapshot,
		RewriteCount:    0,
	}

	if !checkResult.Pass && checkResult.Score < 6 {
		chapter.Status = "coherence_check"
		chapter.RewriteCount = 1
	} else {
		chapter.Status = "done"
	}
	s.db.Create(chapter)
	return chapter, nil
}
```

---

### Task 19: Comic Service

**Files:**
- Create: `internal/service/comic_service.go`

- [ ] **Step 1: 实现漫画生成（单插画 + 多格漫画双路径，含下载和校验）**

```go
package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"novelforge/internal/ai"
	"novelforge/internal/model"
	"gorm.io/gorm"
)

type PanelScript struct {
	Panel    int    `json:"panel"`
	Scene    string `json:"scene"`
	Action   string `json:"action"`
	Dialogue string `json:"dialogue"`
	Camera   string `json:"camera"`
}

type ComicScript struct {
	Panels []PanelScript `json:"panels"`
}

type ComicService struct {
	db       *gorm.DB
	provider ai.Provider
	imageDir string
}

func NewComicService(db *gorm.DB, imageDir string) *ComicService {
	return &ComicService{db: db, imageDir: imageDir}
}

func (s *ComicService) SetProvider(p ai.Provider) { s.provider = p }

func (s *ComicService) Generate(chapter *model.Chapter, novel *model.Novel, outline *model.Outline) error {
	if novel.ImageMode == "single" {
		return s.generateSingle(chapter, novel)
	}
	return s.generateMulti(chapter, novel, outline)
}

func (s *ComicService) generateSingle(chapter *model.Chapter, novel *model.Novel) error {
	const maxPromptLen = 500
	content := chapter.Content
	if len(content) > maxPromptLen {
		content = content[:maxPromptLen]
	}
	prompt := fmt.Sprintf("小说《%s》第%d章插画。根据内容生成一张漫画风格插画：%s",
		novel.Title, chapter.ChapterNo, content)

	resp, err := s.provider.GenerateImage(ai.ImageRequest{Prompt: prompt, Size: "1024x1024", N: 1})
	if err != nil {
		return err
	}
	imagePath, err := s.downloadImage(resp.URLs[0], novel.ID, chapter.ChapterNo, 1)
	if err != nil {
		return err
	}
	urlsJSON, _ := json.Marshal([]string{imagePath})

	page := &model.ComicPage{
		ChapterID:  chapter.ID,
		NovelID:    novel.ID,
		PageNo:     1,
		PanelCount: 1,
		ImageURLs:  string(urlsJSON),
		Status:     "done",
	}
	return s.db.Create(page).Error
}

func (s *ComicService) generateMulti(chapter *model.Chapter, novel *model.Novel, outline *model.Outline) error {
	script, err := s.generateScript(chapter)
	if err != nil {
		return err
	}

	var imagePaths []string
	for _, panel := range script.Panels {
		imgPrompt := fmt.Sprintf("漫画格 %d/%d：%s。动作：%s。镜头：%s。风格：日系黑白漫画。",
			panel.Panel, len(script.Panels), panel.Scene, panel.Action, panel.Camera)
		resp, err := s.provider.GenerateImage(ai.ImageRequest{Prompt: imgPrompt, Size: "1024x1024", N: 1})
		if err != nil {
			return err
		}
		path, err := s.downloadImage(resp.URLs[0], novel.ID, chapter.ChapterNo, panel.Panel)
		if err != nil {
			return err
		}
		imagePaths = append(imagePaths, path)
	}

	// ComicCheck：校验图文匹配
	checkResult, err := s.comicCheck(chapter.Content, script, outline)
	if err == nil && !checkResult.Pass && checkResult.RetryPrompt != "" {
		// 用修正后的 prompt 重新生成（简化：仅一次重试）
		resp, err := s.provider.GenerateImage(ai.ImageRequest{Prompt: checkResult.RetryPrompt, Size: "1024x1024", N: 1})
		if err == nil {
			path, _ := s.downloadImage(resp.URLs[0], novel.ID, chapter.ChapterNo, 0)
			imagePaths = append(imagePaths, path)
		}
	}

	scriptJSON, _ := json.Marshal(script)
	urlsJSON, _ := json.Marshal(imagePaths)
	page := &model.ComicPage{
		ChapterID:  chapter.ID,
		NovelID:    novel.ID,
		PageNo:     1,
		PanelCount: len(script.Panels),
		Script:     string(scriptJSON),
		ImageURLs:  string(urlsJSON),
		Status:     "done",
	}
	return s.db.Create(page).Error
}

func (s *ComicService) generateScript(chapter *model.Chapter) (*ComicScript, error) {
	systemPrompt := `你是专业漫画分镜师。将小说段落转化为 4 格漫画分镜脚本。
输出纯 JSON：{"panels":[{"panel":1,"scene":"场景描述","action":"动作","dialogue":"对话","camera":"特写/中景/远景"}]}`

	const maxContent = 800
	content := chapter.Content
	if len(content) > maxContent {
		content = content[:maxContent]
	}

	resp, err := s.provider.Chat(ai.ChatRequest{
		Messages: []ai.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: "章节内容：" + content},
		},
	})
	if err != nil {
		return nil, err
	}
	var script ComicScript
	json.Unmarshal([]byte(resp.Content), &script)
	if len(script.Panels) == 0 {
		script.Panels = []PanelScript{{Panel: 1, Scene: content, Camera: "中景"}}
	}
	return &script, nil
}

func (s *ComicService) comicCheck(sourceText string, script *ComicScript, outline *model.Outline) (*ComicCheckResult, error) {
	scriptJSON, _ := json.Marshal(script)
	systemPrompt := `你是专业漫画编辑。检查漫画分镜是否准确还原小说场景。
维度：场景还原、人物外观、分镜连贯、画风统一。
输出纯 JSON：{"pass":bool,"issues":["..."],"retry_prompt":"修正后的生图 prompt（如果不通过）"}`

	prompt := fmt.Sprintf("原文段落：%s\n\n人物设定：%s\n\n当前分镜脚本：%s",
		sourceText, outline.CharacterSheets, string(scriptJSON))

	resp, err := s.provider.Chat(ai.ChatRequest{
		Messages: []ai.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}
	var result ComicCheckResult
	json.Unmarshal([]byte(resp.Content), &result)
	return &result, nil
}

func (s *ComicService) downloadImage(url string, novelID uint, chapterNo, pageNo int) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	dir := filepath.Join(s.imageDir, fmt.Sprintf("%d", novelID))
	os.MkdirAll(dir, 0755)
	filename := fmt.Sprintf("%d_%d.png", chapterNo, pageNo)
	fullPath := filepath.Join(dir, filename)

	f, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	io.Copy(f, resp.Body)

	// 返回相对路径，前端可通过 /images/{novel_id}/{filename} 访问
	return fmt.Sprintf("%d/%s", novelID, filename), nil
}
```

---

### Task 19b: OutlineService 补充 GetByNovel

**已在 Task 16 的编辑中补充了 `NewOutlineService` 和 `GetByNovel` 方法（见上文）。**

---

### Task 20: NovelService 主流程协调

**Files:**
- Modify: `internal/service/novel_service.go`（添加 StartGeneration 方法、注入队列和子服务）

- [ ] **Step 1: 更新 NovelService 结构体，注入依赖**

```go
package service

import (
	"encoding/json"
	"novelforge/internal/model"
	"novelforge/internal/task"
	"gorm.io/gorm"
)

type NovelService struct {
	db         *gorm.DB
	outlineSvc *OutlineService
	chapterSvc *ChapterService
	comicSvc   *ComicService
	taskQ      *task.Queue
}

func NewNovelService(db *gorm.DB) *NovelService { return &NovelService{db: db} }

func (s *NovelService) SetOutlineService(svc *OutlineService)   { s.outlineSvc = svc }
func (s *NovelService) SetChapterService(svc *ChapterService)   { s.chapterSvc = svc }
func (s *NovelService) SetComicService(svc *ComicService)       { s.comicSvc = svc }
func (s *NovelService) SetQueue(q *task.Queue)                  { s.taskQ = q }

func (s *NovelService) List(userID uint) ([]model.Novel, error) {
	var novels []model.Novel
	err := s.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&novels).Error
	return novels, err
}

func (s *NovelService) GetByID(novelID uint) (*model.Novel, error) {
	var novel model.Novel
	err := s.db.First(&novel, novelID).Error
	return &novel, err
}

func (s *NovelService) Create(userID uint, title, summary, mode, imageMode string, aiConfigID uint) (*model.Novel, error) {
	novel := model.Novel{
		UserID:     userID,
		Title:      title,
		Summary:    summary,
		Mode:       mode,
		ImageMode:  imageMode,
		AIConfigID: aiConfigID,
	}
	if err := s.db.Create(&novel).Error; err != nil {
		return nil, err
	}
	return &novel, nil
}

func (s *NovelService) StartGeneration(novelID uint) error {
	novel, err := s.GetByID(novelID)
	if err != nil {
		return err
	}

	var outlineInput string
	switch novel.Mode {
	case "blindbox":
		// 盲盒模式：先让 AI 自动选题起名
		titleResp, err := s.outlineSvc.provider.Chat(ai.ChatRequest{
			Messages: []ai.ChatMessage{
				{Role: "system", Content: "你是一个创意小说作者。请自动想一个引人入胜的小说题材和标题。输出 JSON：{\"title\":\"...\",\"summary\":\"一句话梗概\"}"},
			},
		})
		if err != nil {
			return err
		}
		var blindResult struct {
			Title   string `json:"title"`
			Summary string `json:"summary"`
		}
		json.Unmarshal([]byte(titleResp.Content), &blindResult)
		novel.Title = blindResult.Title
		novel.Summary = blindResult.Summary
		s.db.Save(novel)
		outlineInput = blindResult.Summary

	case "outline":
		// 章纲驱动：用户已提供概要，直接用作大纲生成输入
		outlineInput = novel.Summary

	case "inspiration":
		// 灵感起步：用户给了一句话梗概
		outlineInput = novel.Summary

	default:
		outlineInput = novel.Summary
	}

	// 生成大纲
	outline, err := s.outlineSvc.Generate(novel, outlineInput)
	if err != nil {
		return err
	}

	// 解析 chapter_plan
	var chapterPlan []ChapterPlanItem
	json.Unmarshal([]byte(outline.ChapterPlan), &chapterPlan)

	// 将写文任务逐一入队
	for _, cp := range chapterPlan {
		s.taskQ.EnqueueWrite(task.Task{NovelID: novelID, ChapterNo: cp.ChapterNo})
	}

	// 更新小说状态
	novel.Status = "drafting"
	s.db.Save(novel)
	return nil
}
```

- [ ] **Step 2: Task 14 的 NovelService 简化版现在被此完全版替代；Task 14 中 NovelService 仅需保留 Create 和 List 方法，其它扩展移至此处**

---

### Task 20b: 更新 Task Queue 处理函数（写入队列协调逻辑）

**Files:**
- Modify: `main.go`（Task 22 中 taskQ 的初始化逻辑）

在 `main.go` 中定义 writeHandler 和 imageHandler 时需要包含完整的后续调度逻辑：

```go
writeHandler := func(t task.Task) error {
    novel, _ := novelSvc.GetByID(t.NovelID)
    outline, _ := outlineSvc.GetByNovel(t.NovelID)
    var plan []service.ChapterPlanItem
    json.Unmarshal([]byte(outline.ChapterPlan), &plan)
    chapter, err := chapterSvc.Write(t.NovelID, t.ChapterNo, plan, outline)
    if err != nil {
        return err
    }
    // 写文成功后，将生图任务入队（并行：写下一章的同时生当前章的图）
    if chapter.Status == "done" {
        taskQ.EnqueueImage(task.Task{NovelID: t.NovelID, ChapterNo: t.ChapterNo, Type: task.TaskImage})
    }
    return nil
}

imageHandler := func(t task.Task) error {
    novel, _ := novelSvc.GetByID(t.NovelID)
    outline, _ := outlineSvc.GetByNovel(t.NovelID)
    chapter, _ := chapterSvc.GetByNovelAndNo(t.NovelID, t.ChapterNo)
    return comicSvc.Generate(chapter, novel, outline)
}
```

---

### Task 21: SSE 进度推送 + Reader 页面

**Files:**
- Create: `internal/handler/sse.go`
- Create: `internal/handler/chapter.go`
- Create: `templates/reader.html`

- [ ] **Step 1: SSE Handler（含 mutex 并发保护）**

```go
package handler

import (
	"io"
	"sync"
	"github.com/gin-gonic/gin"
)

type SSEHandler struct {
	mu      sync.RWMutex
	clients map[uint]chan string // novelID -> []channel
}

func NewSSEHandler() *SSEHandler {
	return &SSEHandler{clients: make(map[uint]chan string)}
}

func (h *SSEHandler) Subscribe(c *gin.Context) {
	novelID := c.GetUint("novel_id")
	ch := make(chan string, 10)

	h.mu.Lock()
	h.clients[novelID] = ch
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, novelID)
		h.mu.Unlock()
		close(ch)
	}()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	c.Stream(func(w io.Writer) bool {
		msg, ok := <-ch
		if !ok {
			return false
		}
		c.SSEvent("progress", msg)
		return true
	})
}

func (h *SSEHandler) Push(novelID uint, message string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if ch, ok := h.clients[novelID]; ok {
		select {
		case ch <- message:
		default:
			// client too slow, drop message
		}
	}
}
```

- [ ] **Step 2: Chapter Handler（完整实现）**

```go
package handler

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"novelforge/internal/model"
	"novelforge/internal/service"
)

type ChapterHandler struct {
	db  *gorm.DB
	nSvc *service.NovelService
	cSvc *service.ChapterService
}

func NewChapterHandler(db *gorm.DB, nSvc *service.NovelService, cSvc *service.ChapterService) *ChapterHandler {
	return &ChapterHandler{db: db, nSvc: nSvc, cSvc: cSvc}
}

func (h *ChapterHandler) View(c *gin.Context) {
	novelID, _ := strconv.Atoi(c.Param("id"))
	chapterNo, _ := strconv.Atoi(c.Param("no"))

	novel, err := h.nSvc.GetByID(uint(novelID))
	if err != nil {
		c.String(http.StatusNotFound, "作品不存在")
		return
	}

	chapter, err := h.cSvc.GetByNovelAndNo(uint(novelID), chapterNo)
	if err != nil {
		c.String(http.StatusNotFound, "章节不存在")
		return
	}

	var pages []model.ComicPage
	h.db.Where("novel_id = ? AND chapter_id = ?", novelID, chapter.ID).Order("page_no").Find(&pages)

	// 获取上一章/下一章信息
	var prevNo, nextNo int
	h.db.Model(&model.Chapter{}).
		Where("novel_id = ? AND chapter_no < ?", novelID, chapterNo).
		Select("COALESCE(MAX(chapter_no), 0)").Scan(&prevNo)
	h.db.Model(&model.Chapter{}).
		Where("novel_id = ? AND chapter_no > ?", novelID, chapterNo).
		Select("COALESCE(MIN(chapter_no), 0)").Scan(&nextNo)

	c.HTML(http.StatusOK, "reader.html", gin.H{
		"novel":   novel,
		"chapter": chapter,
		"pages":   pages,
		"prevNo":  prevNo,
		"nextNo":  nextNo,
	})
}
```

- [ ] **Step 3: 编写 reader.html**

左侧章节文本（可滚动）+ 右侧漫画展示区（图片网格，`/images/{{.novel_id}}/...` 引用）+ 顶部章节导航（上一章/下一章链接）+ SSE 订阅（`hx-ext="sse" hx-sse="connect:/api/sse?novel_id={{.novel.ID}}"` 监听进度事件）。

---

### Task 22: main.go 全局组装

**Files:**
- Write: `main.go`（整合所有模块）

- [ ] **Step 1: 实现完整的 main.go**

```go
package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
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

	db, _ := gorm.Open(sqlite.Open(cfg.DBPath+"?_journal_mode=WAL"), &gorm.Config{})
	db.AutoMigrate(&model.User{}, &model.AIConfig{}, &model.Novel{}, &model.Outline{}, &model.Chapter{}, &model.ComicPage{})

	// Services（先创建不含 provider 的服务，provider 按请求动态创建）
	authSvc := service.NewAuthService(db)
	outlineSvc := service.NewOutlineService(db)
	chapterSvc := service.NewChapterService(db, outlineSvc)
	comicSvc := service.NewComicService(db, cfg.ImageDir)
	novelSvc := service.NewNovelService(db)

	// SSE Hub
	sseH := handler.NewSSEHandler()

	// Task Queue —— provider 按 novel 的 AI 配置动态创建
	taskQ := task.New(
		// writeHandler
		func(t task.Task) error {
			novel, _ := novelSvc.GetByID(t.NovelID)
			outline, _ := outlineSvc.GetByNovel(t.NovelID)

			// 按作品配置动态创建 provider
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
			sseH.Push(t.NovelID, `{"type":"chapter","chapter_no":`+fmt.Sprint(t.ChapterNo)+`,"status":"done"}`)
			if chapter.Status == "done" {
				taskQ.EnqueueImage(task.Task{NovelID: t.NovelID, ChapterNo: t.ChapterNo, Type: task.TaskImage})
			}
			return nil
		},
		// imageHandler
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
				sseH.Push(t.NovelID, `{"type":"comic","chapter_no":`+fmt.Sprint(t.ChapterNo)+`,"status":"done"}`)
			}
			return err
		},
	)

	novelSvc.SetTaskQueue(taskQ)
	novelSvc.SetOutlineService(outlineSvc)
	novelSvc.SetChapterService(chapterSvc)
	novelSvc.SetComicService(comicSvc)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	novelH := handler.NewNovelHandler(novelSvc, db)
	chapterH := handler.NewChapterHandler(db, novelSvc, chapterSvc)
	aiConfigH := handler.NewAIConfigHandler(db)

	r := gin.Default()
	r.SetTrustedProxies(nil)

	// Session
	store := cookie.NewStore([]byte(cfg.SessionSecret))
	r.Use(sessions.Sessions("novelforge_session", store))

	// 模板
	r.LoadHTMLGlob("templates/*.html")

	// 路由
	r.GET("/login", middleware.GuestOnly(), authH.LoginPage)
	r.POST("/login", middleware.GuestOnly(), authH.Login)
	r.GET("/register", middleware.GuestOnly(), authH.RegisterPage)
	r.POST("/register", middleware.GuestOnly(), authH.Register)
	r.GET("/logout", authH.Logout)

	auth := r.Group("/", middleware.AuthRequired())
	{
		auth.GET("/", novelH.Home)
		auth.POST("/novel", novelH.Create)
		auth.GET("/novel/new", novelH.NewPage)
		auth.POST("/novel/generate", novelH.StartGeneration)
		auth.GET("/novel/:id", novelH.Detail)
		auth.GET("/novel/:id/chapter/:no", chapterH.View)
		auth.GET("/api/sse", sseH.Subscribe)
		auth.GET("/ai-config", aiConfigH.Page)
		auth.POST("/ai-config", aiConfigH.Create)
		auth.DELETE("/ai-config/:id", aiConfigH.Delete)
	}

	// 静态资源
	r.Static("/images", cfg.ImageDir)
	r.Static("/static", "./static")

	r.Run(":" + cfg.Port)
}
```

- [ ] **Step 2: 在 service 包中添加辅助函数 CreateProviderFromConfig**

在任意 service 文件中（如 `novel_service.go`）添加：

```go
import (
	"novelforge/internal/ai"
	"novelforge/internal/model"
)

func CreateProviderFromConfig(cfg *model.AIConfig) ai.Provider {
	return ai.NewProvider(cfg.Provider, cfg.APIKey, cfg.BaseURL, cfg.TextModel, cfg.ImageModel)
}
```

- [ ] **Step 3: NovelService 补充 SetTaskQueue 方法，修正方法名**

```go
func (s *NovelService) SetTaskQueue(q *task.Queue) { s.taskQ = q }
```

- [ ] **Step 4: NovelHandler 补充 Home、Detail、NewPage 和 StartGeneration 方法**

```go
package handler

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"novelforge/internal/model"
	"novelforge/internal/service"
)

type NovelHandler struct {
	svc *service.NovelService
	db  *gorm.DB
}

func NewNovelHandler(svc *service.NovelService, db *gorm.DB) *NovelHandler {
	return &NovelHandler{svc: svc, db: db}
}

func (h *NovelHandler) Home(c *gin.Context) {
	userID := c.GetUint("user_id")
	novels, _ := h.svc.List(userID)
	c.HTML(http.StatusOK, "home.html", gin.H{"novels": novels})
}

func (h *NovelHandler) NewPage(c *gin.Context) {
	var configs []model.AIConfig
	userID := c.GetUint("user_id")
	h.db.Where("user_id = ?", userID).Find(&configs)
	c.HTML(http.StatusOK, "new_novel.html", gin.H{"configs": configs})
}

func (h *NovelHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")
	title := c.PostForm("title")
	mode := c.PostForm("mode")
	imageMode := c.PostForm("image_mode")
	summary := c.PostForm("summary")
	aiConfigID, _ := strconv.Atoi(c.PostForm("ai_config_id"))
	novel, _ := h.svc.Create(userID, title, summary, mode, imageMode, uint(aiConfigID))
	c.Redirect(http.StatusFound, "/novel/"+strconv.Itoa(int(novel.ID)))
}

func (h *NovelHandler) StartGeneration(c *gin.Context) {
	novelID, _ := strconv.Atoi(c.PostForm("novel_id"))
	go h.svc.StartGeneration(uint(novelID))
	c.Redirect(http.StatusFound, "/novel/"+strconv.Itoa(novelID))
}

func (h *NovelHandler) Detail(c *gin.Context) {
	novelID, _ := strconv.Atoi(c.Param("id"))
	novel, _ := h.svc.GetByID(uint(novelID))
	var chapters []model.Chapter
	h.db.Where("novel_id = ?", novelID).Order("chapter_no").Find(&chapters)
	var pages []model.ComicPage
	h.db.Where("novel_id = ?", novelID).Order("chapter_id, page_no").Find(&pages)
	var outline model.Outline
	h.db.Where("novel_id = ?", novelID).Order("version DESC").First(&outline)
	c.HTML(http.StatusOK, "novel_detail.html", gin.H{
		"novel":    novel,
		"chapters": chapters,
		"pages":    pages,
		"outline":  outline,
	})
}
```

---

### Task 23: 验证与编译

- [ ] **Step 1: 完整编译**

```bash
cd /Users/z3/workspace/explore/auto_drama && go build -o novelforge .
```
Expected: 生成 `novelforge` 二进制文件。

- [ ] **Step 2: 启动服务**

```bash
./novelforge
```
Expected: 
```
[GIN-debug] Listening and serving HTTP on :8080
```

- [ ] **Step 3: 浏览器验证**

打开 `http://localhost:8080/register` → 注册 → 登录 → 进入书架 → 添加 AI 配置 → 创建作品。

---

## 实施顺序建议

```
Task 1-5   (脚手架 + 模型 + DB)    ← 无依赖，第一批并行
Task 6-8   (AI Provider)           ← 依赖 Task 2 模型
Task 9      (任务队列)              ← 无依赖
Task 10-12 (Auth 全套)              ← 依赖 Task 3+5
Task 13     (AI Config)             ← 依赖 Task 3+10
Task 14-15 (Novel CRUD + 模板)      ← 依赖 Task 10
Task 16     (Outline Service)       ← 依赖 Task 8
Task 17-18 (Coherence+Chapter)      ← 依赖 Task 16
Task 19     (Comic Service)         ← 依赖 Task 8
Task 20     (NovelService 主流程)    ← 依赖 9+16+18+19
Task 21     (SSE + Reader)           ← 依赖 14+18+19
Task 22-23 (组装 + 验证)             ← 依赖全部以上
```
