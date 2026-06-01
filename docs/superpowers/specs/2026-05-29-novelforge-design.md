# NovelForge 产品设计规格书

> 日期：2026-05-29 | 状态：已确认 | 版本：v1

---

## 一、产品概览

**NovelForge** 是一个 AI 驱动的自动小说写作 + 同步漫画生成工具。用户通过三种模式启动创作，后端异步逐章生成小说正文和配套漫画，实时推送进度。

**核心目标**：最快实现、不过度设计、单机可运行。

### 三种创作模式

| 模式 | 用户输入 | AI 行为 |
|------|---------|---------|
| 灵感起步 | 一句话梗概 | 自动扩写大纲 → 逐章生成正文 + 插图 |
| 章纲驱动 | 详细章节大纲 | 按纲逐章写文 + 配图 |
| 盲盒创作 | 不输入任何内容 | AI 自动选题、起名、生成完整小说+漫画 |

### 两种生图模式

- **单插画**：每章配一张独立插画
- **多格漫画**：生成分镜脚本后，每页 4-6 格带对话气泡的漫画

### 核心质量要求

- 每页漫画情节与对应章节正文严格对应
- 前后章节、前后漫画页之间剧情连贯、画风一致
- 生成过程通过连贯性校验（CoherenceCheck + ComicCheck）保证质量

---

## 二、技术栈

| 层次 | 选型 | 理由 |
|------|------|------|
| 语言 | Go 1.22+ | 后端主力语言 |
| Web 框架 | Gin | 轻量高性能 |
| ORM | GORM + SQLite | 零配置、零外部依赖 |
| 前端 CSS | Tailwind CSS (CDN) | 无构建步骤，零前端工具链 |
| 前端交互 | HTMX + Alpine.js (CDN) | 少量交互，声明式，无需 npm |
| 模板引擎 | Go html/template | 服务端直出，Gin 原生支持 |
| 实时推送 | SSE (Server-Sent Events) | 单向推送生成进度，比 WebSocket 简单 |
| 异步任务 | Go channel 内存队列 | 无需外部消息队列 |
| AI 接入 | 可配置 Provider 插件 | OpenAI / 通义千问 / 自定义兼容接口 |

---

## 三、系统架构

```
┌─────────────────────────────────────────────┐
│                  浏览器                       │
│   Tailwind CSS + HTMX + Alpine.js           │
└──────────────────┬──────────────────────────┘
                   │ HTTP/SSE
┌──────────────────▼──────────────────────────┐
│              Gin HTTP Server                 │
│  ├─ 路由: / /login /register /novel/:id     │
│  │         /novel/:id/chapter/:no           │
│  ├─ 中间件: auth / session / logger          │
│  └─ SSE: 实时推送生成进度                    │
├─────────────────────────────────────────────┤
│              Service Layer                   │
│  ├─ AuthService     用户认证                 │
│  ├─ NovelService    小说创作协调             │
│  ├─ OutlineService  大纲生成                 │
│  ├─ ChapterService  章节写作                 │
│  ├─ ComicService    漫画生成                 │
│  └─ CoherenceCheck  剧情连贯性校验           │
├─────────────────────────────────────────────┤
│              AI Provider Layer               │
│  ├─ interface Provider (Chat, GenerateImage) │
│  ├─ OpenAIDriver    (GPT-4o / DALL·E)       │
│  ├─ QwenDriver     (通义千问)               │
│  └─ CustomDriver   (兼容 OpenAI API 格式)    │
├─────────────────────────────────────────────┤
│              Task Queue (Channel)            │
│  └─ 双队列：写文队列 + 生图队列，并行推进    │
├─────────────────────────────────────────────┤
│              SQLite (GORM)                   │
│  └─ users / ai_configs / novels / outlines   │
│       / chapters / comic_pages               │
└─────────────────────────────────────────────┘
```

**设计要点**：
- 单一二进制文件，`go build` 一条命令出制品
- AI Provider 通过接口注入，新增厂商只加一个文件
- 双队列并行：第 N 章写文的同时，第 N-1 章开始生图
- SSE 实时推送进度，前端 HTMX 监听事件局部更新 DOM

---

## 四、数据库设计

```sql
-- 用户表
users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- AI 配置表（每个用户可配多组）
ai_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,                    -- 配置名称，如 "默认配置"
    provider TEXT NOT NULL,               -- openai / qwen / custom
    api_key TEXT NOT NULL,
    base_url TEXT NOT NULL DEFAULT '',    -- 自定义 API 地址
    text_model TEXT NOT NULL,             -- 文字模型名
    image_model TEXT NOT NULL,            -- 图像模型名
    is_default BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 小说表
novels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    summary TEXT DEFAULT '',
    cover_url TEXT DEFAULT '',            -- 封面图路径
    mode TEXT NOT NULL,                   -- inspiration / outline / blindbox
    image_mode TEXT NOT NULL DEFAULT 'single',  -- single（单插画）/ multi（多格漫画）
    status TEXT NOT NULL DEFAULT 'drafting',  -- drafting / completed / failed
    ai_config_id INTEGER REFERENCES ai_configs(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 大纲表
outlines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    novel_id INTEGER NOT NULL REFERENCES novels(id),
    version INTEGER NOT NULL DEFAULT 1,
    content TEXT NOT NULL,                -- 大纲正文
    character_sheets TEXT DEFAULT '{}',   -- JSON: 人物设定
    world_setting TEXT DEFAULT '{}',      -- JSON: 世界观设定
    chapter_plan TEXT DEFAULT '[]',       -- JSON: 章节规划
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 章节表
chapters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    novel_id INTEGER NOT NULL REFERENCES novels(id),
    chapter_no INTEGER NOT NULL,
    title TEXT NOT NULL,
    content TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',  -- pending/writing/coherence_check/done/failed
    rewrite_count INTEGER DEFAULT 0,
    context_snapshot TEXT DEFAULT '',         -- 生成时使用的上下文快照（调试用）
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(novel_id, chapter_no)
);

-- 漫画页表
comic_pages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chapter_id INTEGER NOT NULL REFERENCES chapters(id),
    novel_id INTEGER NOT NULL REFERENCES novels(id),
    page_no INTEGER NOT NULL,
    panel_count INTEGER NOT NULL DEFAULT 4,   -- 每页几格
    script TEXT DEFAULT '{}',                -- JSON: 分镜脚本
    image_urls TEXT DEFAULT '[]',            -- JSON: 图片文件路径列表
    style_desc TEXT DEFAULT '',              -- 画风描述
    status TEXT NOT NULL DEFAULT 'pending',   -- pending/generating/check/done/failed
    retry_count INTEGER DEFAULT 0,
    context_snapshot TEXT DEFAULT '',         -- 生成上下文快照（调试用）
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**设计原则**：
- JSON 字段（character_sheets、chapter_plan、script、image_urls）用 TEXT 存储，灵活不僵化
- 漫画图片不存入 DB，只存文件路径，文件存本地 `data/images/{novel_id}/`
- context_snapshot 保存生成时的上下文窗口，方便调试和重试
- 每个 novel 绑定一个 ai_config_id，不同作品可用不同 AI 配置

---

## 五、核心数据流

```
用户输入梗概/章纲/留空
         │
         ▼
   盲盒模式? ──Y→ [0] NovelService 自动生成标题 + 故事前提
         │ N
         ▼
   [1] OutlineService → 生成完整故事大纲
       含：人物设定、世界观、章节划分
         │
         ▼
   [2] 逐章循环 ──────────────────────┐
         │                            │
         ▼                            │
   [3] ChapterService                 │
       传入: 大纲 + 前N章摘要         │
       产出: 本章正文                 │
         │                            │
         ▼                            │
   [4] CoherenceCheck ────────────────┤
       检查: 人物一致性、情节逻辑、    │
             世界观冲突               │
       通过? ──N→ 重写本章 ──────────┘
         │ Y
         ▼
   [5] 判断 image_mode
         │
   single│         │multi
         ▼         ▼
   单插画路径      多格漫画路径
   ImageGen     → ComicService（生成分镜脚本）
   一次调用生   → ImageGen（逐格调用 AI）
   成全页插画   → ComicCheck（分镜校验 + 重试）
         │         │
         ├────┐    │
         ▼    │    ▼
   [8] SSE 推送: 第N章完成 ◄──────────┘
         │
         ▼
   循环至全部章节完成
```

### 任务队列协调机制

**双队列设计**：

| 队列 | 作用 | 容量 |
|------|------|------|
| writeQueue | 章节写作 + 连贯性校验 | buffered channel 10 |
| imageQueue | 漫画/插画生成 | buffered channel 10 |

**Worker 数量**：
- writeQueue：1 个 worker（写文需严格顺序，前文完成后才能写下一章）
- imageQueue：1 个 worker（顺序生成，保证前后漫画页风格参照一致）

**队列间依赖**：
- 第 N 章写文完成 → 第 N 章生图任务入队 + 第 N+1 章写文任务入队
- 第 1 章写文完成时先入 imageQueue 生图，再入 writeQueue 写第 2 章（并行）
- writeQueue 的 worker 写完第 N 章后不进入第 N+1 章，直到 imageQueue 中第 N 章开始生图

**降级策略**：
- 章节写入失败 → 标记章节为 failed，继续下一章
- 图像生成失败 → 重试 3 次，仍失败则标记为 failed，跳过继续
- 任一队列满 → 阻塞等待，防止内存膨胀

---

## 六、关键 Prompt 设计原则

### CoherenceCheck（剧情连贯性校验）

```
系统角色：你是一个专业的小说编辑，负责审查章节的剧情连贯性。

任务：检查「当前章节」是否与「前文」在以下维度保持一致：
1. 人物一致性：性格、关系、已知经历是否前后矛盾
2. 情节逻辑：事件因果链是否断裂
3. 世界观一致：设定（魔法体系、等级规则等）是否冲突

输入：
- 前 3 章摘要（拼接文本）：{{previous_summaries}}
- 人物设定：{{character_sheets}}
- 世界观：{{world_setting}}
- 当前章节：{{chapter_content}}

输出 JSON：
{
  "pass": true/false,
  "score": 1-10,
  "issues": ["具体问题描述"],
  "suggestion": "修改建议"
}
```

### ComicCheck（图文匹配校验）

```
系统角色：你是一个专业的漫画编辑，负责审查漫画页是否准确还原了小说场景。

任务：检查漫画分镜是否与原文段落在以下维度一致：
1. 场景还原：关键情节/动作/对话是否缺失
2. 人物外观：角色形象是否与人物设定一致
3. 分镜连贯：前后格画面之间的动作/视角是否流畅
4. 画风统一：与前几页漫画风格是否一致

输入：
- 原文段落：{{source_text}}
- 人物设定：{{character_sheets}}
- 当前分镜脚本：{{comic_script}}
- 前几页漫画描述：{{previous_comic_descriptions}}

输出 JSON：
{
  "pass": true/false,
  "issues": ["具体问题"],
  "retry_prompt": "修正后的生图 prompt"
}
```

### 上下文窗口管理

每次生成维护滑动窗口，包含：
- 前 3 章摘要（每章约 200 字精简）
- 前 3 页漫画描述（每页约 100 字画面描述）
- 当前章节/段落的完整内容

窗口大小可配置，防止 token 超限。

---

## 七、前端页面结构

采用 HTMX 驱动的多页面应用，共 5 个页面：

| 路由 | 页面 | 说明 |
|------|------|------|
| `/login` | 登录页 | 用户名+密码表单 |
| `/register` | 注册页 | 用户名+密码+确认密码 |
| `/` | 书架首页 | 卡片网格展示所有作品，新建按钮 |
| `/novel/:id` | 作品详情页 | 大纲树 + 章节列表 + 进度条 |
| `/novel/:id/chapter/:no` | 阅读页 | 左侧正文 + 右侧漫画，翻页 |

### 交互方案

| 操作 | 实现方式 |
|------|---------|
| 页面跳转 | `<a href>` 链接 |
| 表单提交 | HTMX `hx-post`，局部刷新 |
| 新建作品弹窗 | Alpine.js `x-show` |
| 生成进度 | SSE + HTMX 监听事件 |
| 大纲/章节折叠 | Alpine.js `x-show` |
| Toast 通知 | Alpine.js + 定时器 |

### CDN 依赖（零构建）

```
<script src="https://cdn.tailwindcss.com"></script>
<script src="https://unpkg.com/htmx.org@1.9.10"></script>
<script src="https://unpkg.com/alpinejs@3.13.5"></script>
```

### 布局方案

- **书架首页**：顶部导航 + 新建按钮 + 作品卡片网格（每卡显示封面/标题/模式标签/进度）
- **作品详情页**：返回按钮 + 标题 + 状态条 + 左侧大纲树 + 右侧章节/漫画缩略图网格
- **阅读页**：章节目导航 + 左右分栏（左侧文本区 + 右侧漫画展示区 + 翻页控件）

---

## 八、项目目录结构

```
auto_drama/
├── main.go                    # 入口：组装依赖、启动 server
├── go.mod
├── go.sum
├── .env.example               # 环境变量模板
├── data/                      # 运行时数据（SQLite + 图片）
│   ├── novelforge.db
│   └── images/{novel_id}/{chapter_no}_{page_no}.png
├── internal/
│   ├── config/                # 配置加载（环境变量 → struct）
│   ├── handler/               # Gin 路由处理器（HTTP 层，仅做参数解析和响应）
│   │   ├── auth.go            # /login /register /logout
│   │   ├── novel.go           # /novel CRUD + 创建任务
│   │   ├── chapter.go         # /chapter 列表 + 详情
│   │   ├── comic.go           # 漫画页相关
│   │   └── ai_config.go       # AI 配置 CRUD
│   ├── middleware/             # auth / session / logger
│   ├── model/                 # GORM 数据模型
│   │   ├── user.go
│   │   ├── ai_config.go
│   │   ├── novel.go
│   │   ├── outline.go
│   │   ├── chapter.go
│   │   └── comic_page.go
│   ├── service/               # 核心业务逻辑
│   │   ├── auth_service.go
│   │   ├── novel_service.go   # 协调生成流程
│   │   ├── outline_service.go # 大纲生成 + LLM 调用
│   │   ├── chapter_service.go # 章节写作 + LLM 调用
│   │   ├── comic_service.go   # 分镜生成 + 逐格生图
│   │   └── coherence_check.go # 连贯性校验
│   ├── ai/                    # AI Provider 接口 + 实现
│   │   ├── provider.go        # interface{ Chat(), GenerateImage() }
│   │   ├── openai.go
│   │   ├── qwen.go
│   │   └── custom.go          # 兼容 OpenAI API 格式
│   └── task/                  # 异步任务队列
│       └── queue.go           # 内存 channel 队列 + worker
├── templates/                 # Go HTML 模板
│   ├── layout.html            # 基础布局（含 CDN 引用 + nav）
│   ├── login.html
│   ├── register.html
│   ├── home.html              # 书架首页
│   ├── novel_detail.html      # 作品详情页
│   └── reader.html            # 阅读页（文+漫画）
└── static/                    # 静态资源（可选）
    └── logo.svg
```

**架构约束**：
- `internal/` 下所有包不对外暴露，`main.go` 只管依赖组装
- handler → service → model 单向依赖，service 不 import handler
- AI Provider 通过接口 `ai.Provider` 注入，新增 provider 只加一个文件
- 所有 LLM/图像 API 调用走 task queue 异步处理，不阻塞 HTTP 请求

---

## 九、运行与部署

```bash
# 开发运行
cp .env.example .env
# 编辑 .env 填入 AI API Key
go run main.go
# 浏览器打开 http://localhost:8080

# 编译单文件
go build -o novelforge .
./novelforge
```

**环境变量**（`.env`）：
```
PORT=8080
DB_PATH=data/novelforge.db
IMAGE_DIR=data/images
SESSION_SECRET=change-me
```

**AI 配置通过 Web 界面管理**，不写在环境变量中。`.env` 仅存服务级配置。

---

## 十、非功能需求

- **安全性**：密码 bcrypt 哈希存储，Session cookie HttpOnly
- **错误处理**：生成失败自动重试（最多 3 次），失败后标记状态
- **可观测性**：Gin 内置日志中间件 + 关键节点 log 记录
- **数据持久化**：SQLite WAL 模式，异常退出不丢数据

---

## 十一、弃用的设计决策

以下内容明确不做，避免过度设计：

- ❌ 不使用 PostgreSQL / Redis / 消息队列（SQLite + Channel 够用）
- ❌ 不使用 React / Vue / SPA 框架（HTMX + Alpine.js 够用）
- ❌ 不使用 npm / webpack / vite 构建工具（CDN 直引够用）
- ❌ 不实现 WebSocket（SSE 够用）
- ❌ 不实现 OAuth / 第三方登录（简单账号密码够用）
- ❌ 不实现 Docker（等真正需要再补）
- ❌ 不实现多租户 / RBAC 权限体系
