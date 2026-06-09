# Text-to-Image Agent Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 `auto_drama` 中落地一个可交互的文生图 agent 链路，同时把 agent / workflow / trace / feedback 等 AI 工程能力真实沉淀到项目中，并明确吸收 huobao 两套项目里更成熟的 agent 与多模态设计经验。

**Architecture:** 保留当前 `auto_drama` 的 Go API + Vue SPA + SSE 轻量主仓形态，不直接迁移 huobao 整体架构。新增一套围绕 `ImageJob` 的领域对象、接口、事件和 workflow，将产品入口与通用 `AgentTask` 解耦；在实现过程中，把每个关键设计点都映射到对应的 AI 工程问题与面试问题，并在计划中强制加入对本地 huobao 三仓与开源 huobao 的参考检查点，避免只沿着当前仓库局部最优前进。

**Tech Stack:** Go, Gin, GORM, SQLite, Vue 3, Vite, Axios, EventSource, existing AgentTask/Checkpoint models, existing SSE event transport

---

## Working Rules

本计划除了实现任务，还附带三条执行约束，用来规范后续 AI 辅助开发行为。

1. 每次新增接口或状态机时，都必须回答：这部分是 agent step、tool step 还是 deterministic workflow step。
2. 每次设计 prompt、trace、checkpoint、feedback 时，都必须检查本地 huobao 三仓与开源 huobao 是否已有更成熟的设计可借鉴，不能只按当前仓库最小改。
3. 每完成一个任务组，都要补一个“面试映射”复盘：这部分实践后续能回答什么 AI 工程问题。

## Source Documents

- Spec: `docs/superpowers/specs/2026-06-08-text-to-image-agent-design.md`
- Analysis: `docs/superpowers/analysis/2026-06-08-platform-comparison.md`
- Analysis: `docs/superpowers/analysis/2026-06-08-ai-engineer-evolution-roadmap.md`
- Analysis: `docs/superpowers/analysis/2026-06-08-ai-engineer-interview-bottom-cards.md`
- Historical completed phase: `docs/superpowers/plans/2026-06-05-frontend-streaming-experience-implementation.md`

## File Map

### Backend Domain / Persistence

- Create: `internal/model/image_job.go`
  - 文生图任务对象；承接产品任务状态，不直接等同于底层 `AgentTask`。
- Create: `internal/model/image_prompt_draft.go`
  - prompt 草稿与版本对象；支持人工确认与复盘。
- Create: `internal/model/image_result.go`
  - 图片生成结果、评分与反馈。
- Create: `internal/model/agent_trace_step.go`
  - 执行轨迹步骤对象；支撑 tracing、调试、面试复盘。

### Backend Services / Workflow

- Create: `internal/service/image_job_service.go`
  - 领域服务；创建 job、聚合详情、写入 prompt/result/feedback。
- Create: `internal/service/image_workflow_service.go`
  - 最小 workflow 编排；串联 story analysis、prompt generation、provider image generation、result review。
- Create: `internal/service/image_tools.go`
  - 明确 tool schema 与调用边界；不要把 tool 隐式散落在 handler/service 中。
- Modify: `internal/service/agent_service.go`
  - 扩展与 `ImageJob` 对齐的状态与 trace 写入入口；不要破坏已有通用 task 逻辑。
- Modify: `internal/service/task_handler.go`
  - 若复用现有任务执行器，明确接入 image job workflow 的事件推送与状态更新。

### Backend Handlers / Routes

- Modify: `internal/app/routes.go`
  - 注册新的 image job 领域接口；保留通用 agent 接口作为底层能力。
- Create: `internal/handler/image_job.go`
  - 领域 API handler；只做参数校验、调用 service、返回快照。
- Modify: `internal/handler/sse.go`
  - 继续复用 `/sse` 通道，但支持新的 image job 业务事件。

### Backend Event Contract

- Modify: `internal/task/events.go`
  - 增加 image job 相关事件类型与 payload helper。

### Frontend API / State / Views

- Modify: `web/src/api/index.js`
  - 新增 image job API、feedback API、prompt confirm API，并扩展 SSE 事件解析。
- Create: `web/src/composables/useImageJobStream.js`
  - 可选；如复用 `useNovelStream` 不合适，则单独管理 image job 事件分发。
- Modify: `web/src/views/ChapterView.vue`
  - 增加文生图发起入口与当前章节图像任务区。
- Create: `web/src/components/ImageJobPanel.vue`
  - job 状态、trace 摘要、最近结果面板。
- Create: `web/src/components/PromptDraftPanel.vue`
  - prompt 候选、确认/编辑 UI。
- Create: `web/src/components/ImageResultFeedbackPanel.vue`
  - 结果评分与反馈 UI。

### Tests

- Create: `internal/service/image_job_service_test.go`
- Create: `internal/service/image_workflow_service_test.go`
- Create: `internal/handler/image_job_test.go`
- Modify or Create: `internal/task/events_test.go`
- Optional frontend tests only if repo already has pattern; otherwise verify with `npm run build`.

### Documentation / Learning Output

- Create: `docs/superpowers/analysis/2026-06-08-text-to-image-agent-interview-mapping.md`
  - 每个已落地接口和设计，对应的 AI 工程问题、面试问题、踩坑点。

## Task 1: Lock The Reference Baseline From huobao Before Writing Code

**Files:**
- Read: `/Users/z3/workspace/explore/splay/huobao-drama-backend/**`
- Read: `/Users/z3/workspace/explore/splay/huobao-drama-agent/**`
- Read: `/Users/z3/workspace/explore/huobao-drama/backend/**`
- Create: `docs/superpowers/analysis/2026-06-08-text-to-image-agent-reference-notes.md`

- [ ] **Step 1: Summarize local huobao agent/runtime patterns**

记录至少以下内容：
- supervisor / tools 如何分层
- checkpoint / resume 如何建模
- provider / media service 如何抽象

- [ ] **Step 2: Summarize open-source huobao productized patterns**

记录至少以下内容：
- agent type split
- skills / config driven prompt
- multimodal adapter registry

- [ ] **Step 3: Write explicit “adopt / not adopt now” notes**

在参考文档里按三列写清：
- 当前阶段立即吸收
- 当前阶段保留思路但暂不实现
- 当前阶段明确不做

- [ ] **Step 4: Add a learning checkpoint section**

写出当前阶段你应该能回答的三个问题：
- 为什么不用直接迁移 huobao 架构
- 为什么先吸收 agent 思想而不是整体迁移
- 当前文生图阶段最值得吸收的是哪些设计

- [ ] **Step 5: Verify the reference note is enough to constrain later design**

Expected: 后续任何接口或 workflow 设计，都能在该文档中找到“借鉴来源”或“刻意不借鉴”的理由。

## Task 2: Add New Domain Models For AI Engineering Objects

**Files:**
- Create: `internal/model/image_job.go`
- Create: `internal/model/image_prompt_draft.go`
- Create: `internal/model/image_result.go`
- Create: `internal/model/agent_trace_step.go`
- Modify: `internal/app/app.go`
- Test: `internal/service/image_job_service_test.go`

- [ ] **Step 1: Write failing tests for basic create/list/get behavior**

覆盖最小行为：
- image job 可创建
- prompt draft 可按 version 保存
- image result 可附着到 job
- trace step 可按 step_no 排序读取

- [ ] **Step 2: Run backend tests to verify they fail**

Run: `go test ./internal/service ./internal/model ./internal/app`
Expected: FAIL with missing models or missing migration wiring

- [ ] **Step 3: Implement the four new models**

要求：
- 所有状态字段使用明确字符串枚举
- 保留 `agent_task_id` 与 `ImageJob` 的关联
- 为 prompt/result/trace 预留结构化字段，不用先做大 JSON blob

- [ ] **Step 4: Wire migrations into app bootstrap**

在 `internal/app/app.go` 的迁移列表里加入新模型。

- [ ] **Step 5: Re-run tests**

Run: `go test ./internal/service ./internal/model ./internal/app`
Expected: PASS

- [ ] **Step 6: Write the interview mapping note for this task**

在 `2026-06-08-text-to-image-agent-interview-mapping.md` 里补：
- 为什么 `ImageJob` 不等同于 `AgentTask`
- 为什么 prompt / result / trace 要成为独立对象

## Task 3: Define Tool Contracts And Workflow Boundaries First

**Files:**
- Create: `internal/service/image_tools.go`
- Create: `internal/service/image_workflow_service.go`
- Test: `internal/service/image_workflow_service_test.go`
- Update: `docs/superpowers/analysis/2026-06-08-text-to-image-agent-interview-mapping.md`

- [ ] **Step 1: Write failing tests for tool schemas and workflow step ordering**

至少覆盖：
- `analyze_story`
- `build_image_prompt`
- `generate_image`
- `evaluate_image_result`

- [ ] **Step 2: Document deterministic vs agent steps in test names/comments**

让测试直接表达边界判断：
- 文本理解步骤属于 agent step
- provider 调用与持久化属于 deterministic step

- [ ] **Step 3: Implement minimal input/output structs for each tool**

要求：
- 不接受 `map[string]any` 作为主输入输出接口
- 每个 tool 都有清晰 struct
- 错误分类至少区分 retryable / terminal

- [ ] **Step 4: Implement workflow orchestration skeleton**

第一版只串联步骤，不接真实 provider 也可以，但必须把 step boundary 写清。

- [ ] **Step 5: Run focused tests**

Run: `go test ./internal/service -run 'ImageWorkflow|ImageTool' -v`
Expected: PASS

- [ ] **Step 6: Write learning checkpoint notes**

补充回答：
- 为什么某些步骤用 agent，某些步骤不用
- tool 为什么要强结构化
- 为什么现在不直接上多 agent

## Task 4: Add The Image Job Domain API Layer

**Files:**
- Create: `internal/handler/image_job.go`
- Modify: `internal/app/routes.go`
- Create: `internal/handler/image_job_test.go`
- Modify: `internal/service/image_job_service.go`

- [ ] **Step 1: Write failing handler tests for the first five APIs**

覆盖：
- `POST /novels/:id/chapters/:no/image-jobs`
- `GET /image-jobs/:job_id`
- `GET /image-jobs/:job_id/prompts`
- `POST /image-jobs/:job_id/prompts/:prompt_id/confirm`
- `POST /image-results/:id/feedback`

- [ ] **Step 2: Run tests to verify 404/route failures**

Run: `go test ./internal/handler -run 'ImageJob' -v`
Expected: FAIL with missing routes or handlers

- [ ] **Step 3: Implement request/response DTOs with strict binding**

要求：
- request 中保留 `checkpoint_mode`
- request 中保留 `require_prompt_confirmation`
- response 明确当前 step / current status / waiting checkpoint 状态

- [ ] **Step 4: Register routes without disturbing existing novel/chapter APIs**

新增建议路由：
- `POST /novels/:id/chapters/:no/image-jobs`
- `GET /image-jobs/:job_id`
- `GET /image-jobs/:job_id/prompts`
- `POST /image-jobs/:job_id/prompts/:prompt_id/confirm`
- `POST /image-results/:id/feedback`

- [ ] **Step 5: Make image job API the product-facing entrypoint**

禁止前端主流程直接依赖 `/agent/tasks` 创建 image generation。

- [ ] **Step 6: Re-run handler tests**

Run: `go test ./internal/handler -run 'ImageJob' -v`
Expected: PASS

- [ ] **Step 7: Update interview mapping doc**

补充：
- 领域 API vs runtime API 分层
- 为什么创建任务接口要同步返回 job snapshot

## Task 5: Add Structured Trace And Feedback Persistence

**Files:**
- Modify: `internal/service/image_job_service.go`
- Modify: `internal/service/agent_service.go`
- Create or Modify: `internal/service/image_job_service_test.go`

- [ ] **Step 1: Write failing tests for trace recording and feedback storage**

至少覆盖：
- trace step 按顺序写入
- 失败步骤可记录 error_message
- image result feedback 可写入评分与标签

- [ ] **Step 2: Implement trace write helpers**

要求：
- 统一入口记录 `step_type` / `step_name` / `duration_ms`
- 不允许散落直接写 DB

- [ ] **Step 3: Implement feedback write helpers**

要求：
- rating 有明确范围
- feedback label 预留枚举化空间

- [ ] **Step 4: Re-run service tests**

Run: `go test ./internal/service -run 'ImageJob|Feedback|Trace' -v`
Expected: PASS

- [ ] **Step 5: Add learning checkpoint notes**

补充回答：
- tracing 为什么不能只靠日志
- feedback 为什么必须结构化

## Task 6: Extend SSE Event Contract For Image Jobs

**Files:**
- Modify: `internal/task/events.go`
- Modify: `internal/handler/sse.go`
- Modify: `internal/service/image_workflow_service.go`
- Test: `internal/task/events_test.go`

- [ ] **Step 1: Write failing tests for new event types**

新增事件至少包括：
- `image_job_status`
- `prompt_candidate`
- `checkpoint_required`
- `image_result`
- `agent_trace_step`
- `image_feedback_recorded`

- [ ] **Step 2: Keep fixed SSE transport event name unchanged**

继续保持 `progress` 传输事件名，只扩展业务 `type`。

- [ ] **Step 3: Implement payload helper constructors**

不要在 workflow/service 中手拼 map；统一 helper。

- [ ] **Step 4: Emit events at each workflow milestone**

至少覆盖：
- job created
- planning started
- prompt candidate ready
- waiting confirmation
- image generated
- feedback recorded
- failed

- [ ] **Step 5: Re-run tests and backend build**

Run: `go test ./internal/task -v && go build ./...`
Expected: PASS

- [ ] **Step 6: Update interview mapping doc**

补充：
- 快照和事件各自的职责
- 为什么 trace 事件和 log 事件不该混用

## Task 7: Build The First Product UI For Interactive Text-to-Image

**Files:**
- Modify: `web/src/api/index.js`
- Create or Modify: `web/src/composables/useImageJobStream.js`
- Modify: `web/src/views/ChapterView.vue`
- Create: `web/src/components/ImageJobPanel.vue`
- Create: `web/src/components/PromptDraftPanel.vue`
- Create: `web/src/components/ImageResultFeedbackPanel.vue`

- [ ] **Step 1: Extend frontend API layer with image job endpoints**

加入：
- create job
- get job detail
- list prompts
- confirm prompt
- submit feedback

- [ ] **Step 2: Add SSE event parsing for new business types**

要求：
- `type` 仍是唯一 dispatch key
- `log` 仍不能当状态真相源

- [ ] **Step 3: Add chapter page image generation entry**

用户至少可以：
- 发起文生图
- 看到当前 job 状态
- 看到 prompt 候选
- 进行 prompt 确认

- [ ] **Step 4: Add result list and feedback entry**

用户至少可以：
- 查看生成图片
- 标记好/差
- 提交反馈标签

- [ ] **Step 5: Run frontend build verification**

Run: `npm run build`
Expected: PASS

- [ ] **Step 6: Add learning checkpoint notes**

补充回答：
- 为什么前端要显式暴露 prompt、trace、feedback
- 为什么这不只是结果展示，而是 AI 工程能力展示面

## Task 8: Add Prompt Confirmation And Checkpoint Behavior

**Files:**
- Modify: `internal/service/image_workflow_service.go`
- Modify: `internal/service/agent_service.go`
- Modify: `internal/handler/agent.go` only if reuse is necessary
- Test: `internal/service/image_workflow_service_test.go`

- [ ] **Step 1: Write failing tests for prompt confirmation pause/resume**

至少覆盖：
- `require_prompt_confirmation=true` 时暂停
- confirm 后继续 generate step
- reject / edit 时产生新 prompt draft 或重进 planning

- [ ] **Step 2: Reuse existing Checkpoint model carefully**

要求：
- 不要新起一套重复 checkpoint 机制
- image prompt confirmation 优先落在现有 `Checkpoint` 模型上

- [ ] **Step 3: Implement minimal checkpoint flow**

让 image job 与 agent task / checkpoint 状态联动，但不要一次做完完整 resume 平台。

- [ ] **Step 4: Re-run focused tests**

Run: `go test ./internal/service -run 'Checkpoint|PromptConfirm|ImageWorkflow' -v`
Expected: PASS

- [ ] **Step 5: Write huobao comparison note for this task**

明确记录：
- 当前 checkpoint 设计吸收了 huobao 哪些思想
- 哪些 huobao 的完整 resume 能力当前故意不做

## Task 9: Add Minimal Evaluation And Regression Assets

**Files:**
- Create: `docs/superpowers/analysis/2026-06-08-text-to-image-eval-cases.md`
- Optional Create: `testdata/image_jobs/` if repo patterns allow
- Modify: `internal/service/image_job_service.go` or helper files as needed

- [ ] **Step 1: Define 5-10 representative evaluation cases**

至少覆盖：
- 角色清晰
- 场景清晰
- 容易 prompt 漂移
- 容易角色不一致
- 容易抽象失焦

- [ ] **Step 2: Define feedback labels and rating guidance**

例如：
- `character_inconsistent`
- `scene_mismatch`
- `style_off`
- `too_generic`
- `prompt_overfit`

- [ ] **Step 3: Ensure API/UI can carry these labels without new schema churn**

Expected: 不需要再次大改模型即可支持后续评估扩展。

- [ ] **Step 4: Add interview mapping note**

补充：
- 为什么最小 eval 需要尽早做
- 如何从 demo 走向可迭代系统

## Task 10: Write The Execution Retrospective And Interview Mapping

**Files:**
- Update: `docs/superpowers/analysis/2026-06-08-text-to-image-agent-interview-mapping.md`

- [ ] **Step 1: Summarize each shipped interface and its AI engineering lesson**

至少按以下表结构写：
- 接口 / 事件 / 模型
- 训练的 AI 工程能力
- 能回答的面试问题
- 借鉴自 huobao 的点

- [ ] **Step 2: Write down 3-5 real failure modes observed during implementation**

例如：
- prompt output shape drift
- checkpoint state mismatch
- trace incomplete on failure
- provider errors not classified correctly

- [ ] **Step 3: Write down what was deliberately not implemented yet**

例如：
- 多 agent 协作
- 独立 agent 服务化
- 完整 resume 平台

- [ ] **Step 4: Make the retrospective reusable for interview prep**

Expected: 文档可以直接作为后续面试题准备提纲。

## Verification

- Run: `go test ./...`
- Run: `go build ./...`
- Run: `npm run build`

## Out of Scope For This Plan

- 完整短剧平台对象迁移
- 视频生成与合成主链路
- 独立 agent gRPC 服务化
- 通用多 agent 平台化
- 全量 RAG 系统

## Success Criteria

完成本计划后，应满足以下条件：

1. 用户可以从章节页发起一条清晰可见的文生图任务。
2. 前端可以观察到 prompt 生成、等待确认、结果生成、反馈提交的全过程。
3. 后端存在明确的 `ImageJob` / prompt / result / trace 数据结构。
4. 项目中真实体现出 agent / workflow / trace / feedback 等 AI 工程能力，而不是停留在口头概念。
5. 每个关键设计点都能追溯到“当前项目目标”和“huobao 的可借鉴设计实例”。
