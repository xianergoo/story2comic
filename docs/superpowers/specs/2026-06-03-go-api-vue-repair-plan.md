# NovelForge Go API + Vue 架构修复方案

> 日期：2026-06-03 | 状态：待实施 | 版本：v1

---

## 一、文档目的

本文档用于指导当前项目从"SSR/模板 + Vue 半迁移状态"收口为明确的 **Go API + Vue SPA** 架构，并给出可直接执行的修复顺序、接口契约建议、验证标准和实施边界。

本文档面向实现者，目标不是讨论新功能，而是优先解决当前系统的架构不一致、前后端协议错位、主链路不可用等问题。

---

## 二、最终架构结论

项目应明确收口到以下目标架构：

```text
浏览器（Vue SPA）
  -> /api/v1/**        JSON API
  -> /api/v1/sse       SSE 进度推送
  -> /images/**        本地图片资源

Go 服务（Gin）
  - Session 鉴权
  - API 路由
  - Handler（只返回 JSON / SSE）
  - Service 业务层
  - Task Queue 异步任务
  - AI Provider 层
  - SQLite
```

### 架构原则

- Go 后端只负责 API、SSE、Session、静态资源和异步任务。
- Vue 前端是唯一页面层，负责路由、表单、状态展示和 SSE 订阅。
- 后端禁止再承担模板渲染职责。
- 前后端之间只能通过稳定、统一的 JSON 契约通信。

### 不再采用的方案

- 不再继续保留 Go 模板页面作为主要界面。
- 不再混用 `c.HTML`、`c.Redirect` 与 JSON API。
- 不再同时维护 SSR 和 SPA 两套业务入口。

---

## 三、当前项目的关键问题

### 1. 鉴权链路不完整

当前 `authMiddleware` 只校验 session 中是否存在 `user_id`，但没有将用户 ID 写入 Gin context。多个 handler 通过 `c.GetUint("user_id")` 读取当前用户时，实际可能拿到 `0`，导致：

- 小说列表查询不到当前用户数据。
- 小说创建记录到错误的 `user_id`。
- AI 配置查询、写入、删除可能全部失效。

这是当前最先要修的主问题之一。

### 2. API 路由与 Handler 行为不一致

虽然路由已经统一挂在 `/api/v1` 下，但多个 handler 仍然使用：

- `c.HTML(...)`
- `c.Redirect(...)`
- `c.String(...)`

这与 Vue 前端通过 axios 请求 JSON 的方式完全不匹配，导致：

- 列表接口返回 HTML 而不是 JSON。
- 创建接口返回重定向而不是实体数据。
- 详情接口返回模板而不是对象结构。

### 3. 项目已迁移到 Vue，但 SSR 残留未清理

仓库已具备 `web/` 前端目录和 Vue 路由、页面、axios 封装，但 handler 仍保留服务端模板时代的输入输出逻辑。当前本质上是两套交互模型拼在一起：

- 前端认为自己在调用 JSON API。
- 后端仍把自己当成模板站点。

如果不先做架构收口，继续加功能只会扩大混乱。

### 4. 前后端字段命名和模式枚举不统一

例如创建小说时：

- 前端使用 `prompt`
- 后端读取 `summary`
- 前端盲盒模式使用 `random`
- 后端业务逻辑识别 `blindbox`

这种偏差会直接导致主流程不可用。

### 5. 章节与 AI 配置接口同样停留在表单/模板时代

章节查看和 AI 配置仍是模板式 handler，不适合当前 Vue 架构，必须同步改为 JSON API。

---

## 四、修复目标

实施完成后，系统至少要满足以下目标：

1. 所有页面交互都通过 Vue 完成。
2. 所有业务接口都返回统一 JSON。
3. 登录态能稳定传递到业务 handler。
4. 创建小说主链路可端到端跑通。
5. 详情页可稳定展示小说、章节、进度和控制动作。
6. SSE 可作为详情页增量更新机制使用。
7. 后端构建通过，前端构建通过。

---

## 五、后端边界与职责调整

### 1. Handler 层职责

Handler 只做四件事：

- 接收参数
- 参数校验
- 调用 service
- 返回统一 JSON

Handler 不应再：

- 渲染模板
- 执行页面跳转
- 直接拼装复杂业务流程

### 2. Service 层职责

Service 负责业务编排，包括：

- 用户认证
- 小说创建与生成编排
- 大纲生成
- 章节写作
- 漫画生成
- Agent 任务和检查点逻辑

Service 不应依赖 HTTP 细节，也不应假设前端页面结构。

### 3. App 层职责

`internal/app` 只负责：

- 加载配置
- 初始化数据库
- 装配 service / handler / task queue
- 注册 API 路由
- 注册静态资源路由

后续如果需要产线部署 Vue 打包产物，可在这里增加静态文件服务，但不恢复模板渲染。

---

## 六、统一接口规则

### 1. 响应格式统一

所有 JSON API 应使用统一响应结构：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

错误响应建议保持统一语义：

```json
{
  "code": 4001,
  "message": "参数错误",
  "error": "具体错误信息"
}
```

要求：

- 成功接口统一通过 `v1.Success` / `v1.SuccessWithMessage` 等封装输出。
- 错误接口统一通过 `v1.ValidationError`、`v1.Unauthorized`、`v1.NotFound`、`v1.InternalServerError` 等输出。
- 禁止业务 handler 再直接返回风格不一致的 `gin.H{"error": ...}`。

### 2. 请求体统一为 JSON

在 `Go API + Vue` 架构下，业务写操作统一使用 JSON 请求体：

- `POST`
- `PUT`
- 必要时的 `PATCH`

不再使用：

- `PostForm`
- 表单重定向式交互

### 3. DTO 优先于直接暴露数据库模型

关键接口建议定义请求/响应 DTO，而不是让前端依赖数据库表字段的自然形态。这样能减少后续字段调整对前端的影响。

---

## 七、核心接口契约建议

### 1. 认证接口

#### `POST /api/v1/auth/register`

请求：

```json
{
  "username": "alice",
  "password": "123456",
  "confirm_password": "123456"
}
```

响应：

```json
{
  "code": 0,
  "message": "注册成功",
  "data": {
    "id": 1,
    "username": "alice"
  }
}
```

#### `POST /api/v1/auth/login`

请求：

```json
{
  "username": "alice",
  "password": "123456"
}
```

#### `POST /api/v1/auth/logout`

无请求体。

#### `GET /api/v1/auth/me`

返回当前登录用户基础信息。

### 2. 小说接口

#### `GET /api/v1/novels`

返回当前用户的小说列表，建议列表项包含：

```json
{
  "id": 1,
  "title": "示例小说",
  "summary": "一句话梗概",
  "mode": "inspiration",
  "image_mode": "single",
  "status": "drafting",
  "text_status": "writing",
  "image_status": "idle",
  "chapter_count": 3,
  "created_at": "...",
  "updated_at": "..."
}
```

#### `POST /api/v1/novels`

请求建议统一为：

```json
{
  "title": "示例小说",
  "summary": "一句话梗概",
  "mode": "inspiration",
  "image_mode": "single",
  "ai_config_id": 1
}
```

补充约束：

- `mode` 只允许：`inspiration`、`outline`、`blindbox`
- 禁止继续使用 `random`
- `summary` 作为统一输入字段，禁止继续使用 `prompt`

响应建议返回新建作品的基础信息，前端自行跳转到详情页。

#### `GET /api/v1/novels/:id`

返回详情聚合结构：

```json
{
  "novel": {},
  "outline": {},
  "chapters": [],
  "pages": [],
  "progress": {
    "planned": 10,
    "text_done": 3,
    "comic_done": 1,
    "next_chapter_no": 4
  }
}
```

#### `POST /api/v1/novels/:id/stop`

建议请求体：

```json
{
  "pipeline": "text"
}
```

当前可以先只保证 `text` 流程稳定。

#### `POST /api/v1/novels/:id/resume`

恢复文本生成流程。

#### `POST /api/v1/novels/:id/chapters/:no/regenerate`

请求：

```json
{
  "suggestion": "请让本章节奏更紧凑"
}
```

### 3. AI 配置接口

#### `GET /api/v1/ai-configs`

返回当前用户 AI 配置列表。

#### `POST /api/v1/ai-configs`

请求：

```json
{
  "name": "默认配置",
  "provider": "openai",
  "api_key": "sk-xxx",
  "base_url": "https://example.com/v1",
  "text_model": "gpt-4o-mini",
  "image_model": "dall-e-3",
  "is_default": true
}
```

#### `PUT /api/v1/ai-configs/:id`

更新同样采用 JSON 请求体。

#### `DELETE /api/v1/ai-configs/:id`

删除当前用户的指定配置。

### 4. 章节接口

#### `GET /api/v1/novels/:id/chapters/:no`

返回：

```json
{
  "novel": {},
  "chapter": {},
  "pages": [
    {
      "page_no": 1,
      "panel_count": 4,
      "image_urls": ["/images/..."]
    }
  ],
  "prev_chapter_no": 1,
  "next_chapter_no": 3
}
```

此接口不再返回阅读模板。

---

## 八、状态机收口建议

当前模型中已经存在 `status`、`text_status`、`image_status`，这是正确方向，但必须统一语义。

### 1. Novel.Status

- `drafting`
- `completed`
- `failed`
- `stopped`

### 2. Novel.TextStatus

- `idle`
- `writing`
- `paused`
- `done`
- `failed`

### 3. Novel.ImageStatus

- `idle`
- `generating`
- `paused`
- `done`
- `failed`

### 4. 约束

- 所有 handler、service、task handler 使用同一套枚举语义。
- 不允许在不同文件里自由发明状态字符串。
- 实施阶段建议将这些状态提炼成常量，减少拼写错误。

---

## 九、SSE 设计建议

Vue 详情页需要把 SSE 作为正式机制，而不是可有可无的附属能力。

建议统一消息结构，例如：

```json
{
  "type": "progress",
  "step": "outline",
  "detail": "正在生成大纲"
}
```

```json
{
  "type": "chapter",
  "chapter_no": 3,
  "status": "done"
}
```

```json
{
  "type": "error",
  "chapter_no": 3,
  "msg": "写文失败: xxx"
}
```

前端详情页建议：

1. 首次进入页面时请求详情接口。
2. 同时建立 SSE 订阅。
3. 收到消息后局部更新章节状态、完成数和当前进度。
4. 页面卸载时关闭 SSE 连接。

---

## 十、Vue 前端修复建议

### 1. 统一创建小说字段

前端创建表单应统一为：

```js
const createForm = ref({
  title: '',
  summary: '',
  mode: 'inspiration',
  image_mode: 'single',
  ai_config_id: null,
})
```

必须修正以下不一致：

- `prompt -> summary`
- `random -> blindbox`

### 2. 前端自己控制跳转

创建成功后：

- 后端返回新建小说 `id`
- 前端自行 `router.push(`/novel/${id}`)`

后端不再负责页面跳转。

### 3. 页面优先级

先保证以下页面稳定可用：

- `LoginView.vue`
- `RegisterView.vue`
- `HomeView.vue`
- `NovelDetail.vue`

AI 配置页可以作为第二阶段补齐。

### 4. API 封装统一

建议前端 API 层保持模块化：

- `authAPI`
- `novelAPI`
- `chapterAPI`
- `aiConfigAPI`
- `agentAPI`

并统一处理错误响应，避免各页面自己猜测 `res.data` 结构。

---

## 十一、实施顺序

以下顺序是推荐的最小风险实施路径，必须尽量按顺序执行，避免同时进行大面积重构。

### 阶段 1：修复鉴权主链路

目标：让当前登录用户身份在业务 handler 中稳定可用。

任务：

1. 修复 `authMiddleware`，将 session 中的 `user_id` 写入 Gin context。
2. 为 session 中 `user_id` 的类型做归一化处理。
3. 验证 `GET /api/v1/auth/me` 正常返回当前用户。

### 阶段 2：统一小说相关 handler 为 JSON API

目标：打通最核心的业务主链路。

需要改造：

1. `NovelHandler.Home`
2. `NovelHandler.Create`
3. `NovelHandler.Detail`
4. `NovelHandler.Stop`
5. `NovelHandler.Resume`
6. `NovelHandler.RegenerateChapter`

要求：

- 使用 `ShouldBindJSON`
- 返回统一 JSON
- 不再使用 `HTML`、`Redirect`

### 阶段 3：统一 AI 配置 handler 为 JSON API

目标：保证用户可以在 Vue 中配置 AI Provider。

需要改造：

1. `AIConfigHandler.Page` -> 改为列表接口
2. `AIConfigHandler.Create`
3. `AIConfigHandler.Update`
4. `AIConfigHandler.Delete`

### 阶段 4：统一章节接口为 JSON API

目标：为详情页和未来阅读页提供一致的数据接口。

需要改造：

1. `ChapterHandler.View`

要求：

- 返回章节、漫画页、前后章编号等结构化 JSON
- 不再依赖模板文件

### 阶段 5：修正 Vue 页面字段与请求方式

目标：让前端与后端新契约一致。

任务：

1. `prompt` 改为 `summary`
2. `random` 改为 `blindbox`
3. 创建小说成功后由前端路由跳转
4. 详情页按新 JSON 结构读取数据

### 阶段 6：联调主业务链路

必须实际验证：

1. 注册
2. 登录
3. 获取当前用户
4. 创建 AI 配置
5. 创建小说
6. 查看小说列表
7. 查看小说详情
8. 停止生成
9. 恢复生成
10. 重生成章节

### 阶段 7：接入 SSE 增量更新

目标：详情页实时更新生成进度。

任务：

1. 在详情页建立 SSE 订阅
2. 接收章节完成、进度、错误等事件
3. 局部更新页面状态

---

## 十二、建议的目录职责

建议继续使用现有目录，但职责上做明确收口：

```text
internal/
  app/
    app.go
    routes.go
  handler/
    auth.go
    novel.go
    chapter.go
    ai_config.go
    agent.go
    sse.go
  service/
    auth_service.go
    novel_service.go
    outline_service.go
    chapter_service.go
    comic_service.go
    agent_service.go
    task_handler.go
  task/
    queue.go
    factory.go
    handler.go
  model/
    *.go
  ai/
    provider.go
    openai.go
    qwen.go
    custom.go

web/
  src/
    api/
    router/
    stores/
    views/
    components/
```

说明：

- `templates/` 思路不再继续扩展。
- 默认示例组件如 `HelloWorld.vue` 应清理。
- `web/README.md` 应更新为项目真实开发说明，而不是保留 Vite 初始模板描述。

---

## 十三、验证标准

### 1. 构建验证

后端：

```bash
go build ./...
```

前端：

```bash
npm run build
```

### 2. 接口验证

至少手动验证以下接口：

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`
- `GET /api/v1/novels`
- `POST /api/v1/novels`
- `GET /api/v1/novels/:id`
- `POST /api/v1/novels/:id/stop`
- `POST /api/v1/novels/:id/resume`
- `GET /api/v1/ai-configs`
- `POST /api/v1/ai-configs`

### 3. 联调验证

必须实际走完以下流程：

1. 注册新用户
2. 登录
3. 创建 AI 配置
4. 创建小说
5. 进入详情页
6. 观察章节生成状态变化
7. 执行停止/恢复
8. 检查 SSE 是否推送状态变更

---

## 十四、实施禁区

实施过程中禁止出现以下行为：

1. 继续在业务 handler 中使用 `c.HTML`、`c.Redirect`、`c.String`
2. 继续混用表单请求与 JSON 请求
3. 在没有统一契约前继续追加页面功能
4. 同时兼容旧模板页面和新 Vue 页面
5. 在未打通主链路前优先做 UI 美化
6. 在多个文件中随意发明新的状态值

---

## 十五、给实现者的执行目标

可直接作为实施任务说明：

```text
目标：将当前项目彻底收口为 Go API + Vue 架构。

硬性要求：
1. Go 后端只返回 JSON / SSE / 静态资源，不再渲染 HTML 模板。
2. Vue 前端作为唯一页面层，通过 /api/v1 调用后端。
3. 修复 session 到 gin context 的 user_id 注入问题。
4. 统一 auth / novels / chapters / ai-configs 的请求与响应契约。
5. 打通 register/login -> create novel -> list/detail -> stop/resume 的主链路。
6. 清理 SSR 残留逻辑，避免后端继续使用 c.HTML / c.Redirect / c.String。
7. 在完成后使用 go build ./... 和 npm run build 验证。
```

---

## 十六、结论

当前项目并不是方向错误，而是处于一次未完成的架构迁移中。正确的修复方式不是继续补功能，而是优先完成一次明确的收口：

- 后端只做 Go API
- 前端只做 Vue 页面
- 中间通过统一 JSON 契约衔接

只要先把这条主线打通，后续无论是 Agent、Checkpoint、SSE 细化、漫画页展示还是部署形态，都能在清晰边界上继续演进。

---

## 十七、暂缓优化项

以下事项已识别，但本轮先不继续展开，保留为后续优化清单：

1. 首页创建小说时增加 AI 配置下拉选择，而不是只依赖默认配置。
2. 详情页补充更完整的进度面板，明确展示 planned / text_done / comic_done。
3. AI 配置页将 provider 从自由输入改为受控下拉选项。
4. 后端进一步清理模板时代命名，例如 `Home`、`Page` 等 API 语义不清的方法名。
5. 为 Agent 任务补齐前端页面与交互入口。
6. 将状态字符串继续收口为常量，减少多处散落定义。
7. 对详情页 SSE 做更细粒度的连接重试与断线恢复策略。
