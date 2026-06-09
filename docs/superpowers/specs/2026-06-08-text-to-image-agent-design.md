# Text-to-Image Agent Design For auto_drama

> 日期：2026-06-08 | 状态：草案 | 面向项目：`auto_drama`

## 一、目标

本文档不只是定义“文生图功能怎么做”，而是要把当前 `auto_drama` 的文生图改进，设计成一个既能快速可交互上线、又能系统训练 AI 工程能力的产品链路。

因此本文档同时服务两个目标：

1. 产品目标：先做出可用、可交互的文生图能力
2. 个人目标：把这条链路设计成后续 AI 开发工程师面试中的真实底牌

换句话说，后续新增的接口、事件、状态和数据结构，不应该只是为了“能生成图片”，还应该让这些 AI 工程里的核心问题真正体现在项目设计中：

- agent / workflow 边界判断
- tool calling 契约
- prompt version 管理
- execution trace
- human-in-the-loop
- eval / feedback loop

## 二、当前项目现状

当前 `auto_drama` 已有基础接口和实时能力：

### 1. 现有业务接口

- `GET /api/v1/novels`
- `POST /api/v1/novels`
- `GET /api/v1/novels/:id`
- `POST /api/v1/novels/:id/stop`
- `POST /api/v1/novels/:id/resume`
- `GET /api/v1/novels/:id/chapters/:no`
- `POST /api/v1/novels/:id/chapters/:no/regenerate`

### 2. 现有 agent 接口

- `POST /api/v1/agent/tasks`
- `GET /api/v1/agent/tasks`
- `GET /api/v1/agent/tasks/:task_id`
- `POST /api/v1/agent/tasks/:task_id/cancel`
- `GET /api/v1/agent/tasks/:task_id/checkpoints`
- `PUT /api/v1/agent/checkpoints/:checkpoint_id`

### 3. 现有 SSE 能力

当前统一通过 `GET /api/v1/sse?novel_id=:id` 订阅，并已有这些事件类型：

- `novel_status`
- `progress_summary`
- `chapter_status`
- `chapter_stream`
- `outline_stream`
- `comic_status`
- `log`
- `error`

### 4. 当前不足

当前虽然已有通用 agent 接口，但它更像“平台预埋层”，还不是用户真正使用的产品主链路。对当前文生图目标来说，最大问题有三个：

1. 缺少领域化的文生图接口，用户只能间接走章节生成和漫画链路
2. 缺少将 agent 分析步骤显式暴露给前端的接口与事件
3. 缺少 prompt、tool、trace、feedback 这几类对 AI 工程和面试都极其重要的结构层

## 三、设计原则

## 1. 产品接口优先使用领域 API，而不是直接暴露通用 agent API

也就是说：

- 前端产品主流程不应直接依赖 `POST /agent/tasks`
- 前端应优先调用“文生图领域接口”
- 通用 agent task 接口保留给底层编排和调试用途

这样做的原因是：

1. 产品心智更清楚
2. 后续 agent 内部实现可以替换
3. 你在面试里可以清楚讲出“领域 API”和“通用 agent runtime API”的边界

对应的 AI 工程问题：

- 为什么产品接口不直接暴露底层 agent？
- agent 应该作为业务能力，还是产品协议本身？

## 2. 先用单 agent + deterministic workflow，暂不直接上多 agent

当前阶段最合适的策略不是多 agent，而是：

- 用一个小型 image generation agent 负责理解和规划
- 用 deterministic workflow 执行可枚举步骤
- 对模型调用、图片生成、结果持久化等步骤进行严格控制

这样做的原因是：

1. 当前先做可用链路，不要过早引入复杂度
2. 更适合训练 agent / workflow 边界判断
3. 更容易做 trace、retry、checkpoint

对应的 AI 工程问题：

- 什么时候该用 agent，什么时候该用 workflow？
- 为什么不用多 agent？

## 3. 所有关键节点都必须结构化

后续文生图链路中的以下对象必须结构化，而不是只用文本拼接：

- prompt candidate
- agent decision
- tool input
- tool output summary
- image result metadata
- feedback / rating
- trace step

对应的 AI 工程问题：

- 为什么 structured output 比自然语言输出更适合工程化？
- tool calling 为什么必须结构化？

## 四、建议新增的领域对象

为了不强行改写当前 `Novel / Chapter / ComicPage` 主轴，建议新增一套轻量对象，专门服务当前文生图链路。

## 1. `ImageJob`

表示一次完整的文生图请求。

建议核心字段：

- `id`
- `novel_id`
- `chapter_no`
- `source_type`
- `source_text`
- `status`
- `agent_task_id`
- `prompt_version`
- `selected_prompt_id`
- `requested_count`
- `generated_count`
- `provider_name`
- `model_name`
- `created_at`
- `updated_at`

用途：

- 作为用户可见的“生成任务”对象
- 承接产品接口和前端展示

面试价值：

- 可以回答“为什么要把任务对象单独建模”
- 可以回答“任务和 agent task 的边界怎么划分”

## 2. `ImagePromptDraft`

表示 agent 或 workflow 生成的 prompt 草稿。

建议核心字段：

- `id`
- `image_job_id`
- `version`
- `prompt_text`
- `negative_prompt`
- `style_tags`
- `character_refs`
- `scene_refs`
- `source_summary`
- `status`
- `created_by`

用途：

- 支持 prompt 版本和人工修改
- 支持生成前确认和生成后复盘

面试价值：

- 可以回答“prompt 为什么要版本化”
- 可以回答“prompt 为什么不该散落在代码里”

## 3. `ImageResult`

表示一次图片生成的产物。

建议核心字段：

- `id`
- `image_job_id`
- `prompt_draft_id`
- `provider_name`
- `model_name`
- `image_url`
- `seed`
- `status`
- `error_message`
- `rating`
- `feedback_label`
- `created_at`

用途：

- 支持多图结果管理
- 支持人工评分和后续评估

面试价值：

- 可以回答“如何做 AI 结果评估与反馈闭环”

## 4. `AgentTraceStep`

表示一次 agent / workflow 执行中的结构化步骤。

建议核心字段：

- `id`
- `image_job_id`
- `agent_task_id`
- `step_no`
- `step_type`
- `step_name`
- `input_summary`
- `output_summary`
- `status`
- `duration_ms`
- `error_message`
- `created_at`

用途：

- execution trace
- 调试和面试复盘

面试价值：

- 可以回答“怎么做 tracing”
- 可以回答“系统坏了怎么定位”

## 五、建议新增的接口

## 1. 创建文生图任务

`POST /api/v1/novels/:id/chapters/:no/image-jobs`

### 请求体建议

```json
{
  "source_mode": "chapter_full",
  "selected_text": "",
  "image_count": 4,
  "checkpoint_mode": "essential",
  "require_prompt_confirmation": true,
  "provider_override": "",
  "model_override": "",
  "style_profile": "cinematic_storyboard"
}
```

### 设计意图

这个接口不是“直接生成图片”，而是“创建一次领域任务”。

handler 层只做 deterministic 工作：

- 校验 novel/chapter 是否存在
- 组装 `ImageJob`
- 创建底层 `AgentTask`
- 投递异步执行

agent / workflow 决策发生在任务内部，而不是请求入口。

### 这里训练的 AI 工程能力

- agent / workflow 边界判断
- 领域 API 和底层 runtime API 分层

### 面试中能回答的问题

- 为什么产品接口不直接暴露通用 agent task？
- 为什么创建任务要同步返回 job，而不是等待结果？

## 2. 查询文生图任务详情

`GET /api/v1/image-jobs/:job_id`

### 返回建议

返回以下结构：

- job 基础状态
- 当前步骤
- prompt drafts 摘要
- results 摘要
- 最近 trace 摘要
- 是否等待人工确认

### 设计意图

前端不必通过多个接口拼当前状态，任务详情就是主快照来源。

### 这里训练的 AI 工程能力

- 快照模型设计
- 异步任务状态聚合

### 面试中能回答的问题

- 如何设计 AI 任务详情页的数据模型？
- 为什么实时系统仍然需要快照接口？

## 3. 获取任务执行轨迹

`GET /api/v1/image-jobs/:job_id/traces`

### 返回建议

```json
{
  "items": [
    {
      "step_no": 1,
      "step_type": "agent_decision",
      "step_name": "analyze_story",
      "status": "done",
      "input_summary": "chapter_full",
      "output_summary": "2 characters, 1 scene, 1 prompt candidate",
      "duration_ms": 820
    }
  ]
}
```

### 设计意图

不要把 trace 藏在日志里，而要让它成为正式可查询对象。

### 这里训练的 AI 工程能力

- tracing
- failure mode analysis

### 面试中能回答的问题

- 如何定位 AI workflow 中哪一步不稳定？
- 为什么 execution trace 要产品化？

## 4. 获取 prompt 候选列表

`GET /api/v1/image-jobs/:job_id/prompts`

### 设计意图

让 prompt 成为正式对象，而不是隐式文本。

### 这里训练的 AI 工程能力

- prompt versioning
- structured output

### 面试中能回答的问题

- prompt 为什么需要版本化和独立存储？

## 5. 确认或修改 prompt

`POST /api/v1/image-jobs/:job_id/prompts/:prompt_id/confirm`

### 请求体建议

```json
{
  "action": "confirm",
  "edited_prompt_text": "",
  "edited_negative_prompt": ""
}
```

### 设计意图

这是当前阶段最小 human-in-the-loop 节点。

如果 `require_prompt_confirmation=true`，任务在该点暂停，等待用户确认后再继续调用图片生成 provider。

### 这里训练的 AI 工程能力

- human-in-the-loop
- checkpoint 设计

### 面试中能回答的问题

- 什么情况下需要人工确认？
- 为什么不让 agent 完全自动跑到底？

## 6. 重试任务

`POST /api/v1/image-jobs/:job_id/retry`

### 请求体建议

```json
{
  "retry_scope": "generate_only"
}
```

可选值：

- `generate_only`
- `regenerate_prompt`
- `full`

### 设计意图

必须让重试范围显式化，不能只提供一个模糊的“再来一次”。

### 这里训练的 AI 工程能力

- failure recovery
- workflow step boundary

### 面试中能回答的问题

- AI 任务失败后如何做分层重试？

## 7. 评分与反馈

`POST /api/v1/image-results/:id/feedback`

### 请求体建议

```json
{
  "rating": 4,
  "feedback_label": "character_inconsistent",
  "comment": "角色脸型和设定不一致"
}
```

### 设计意图

让人工反馈成为正式数据，而不是口头评价。

### 这里训练的 AI 工程能力

- minimal eval design
- human feedback loop

### 面试中能回答的问题

- 你如何建立最小评估闭环？

## 六、建议新增的 SSE 事件

当前仍然可以复用 `GET /api/v1/sse?novel_id=:id` 作为传输通道，但需要新增业务事件类型。

## 1. `image_job_status`

表示文生图任务整体状态。

示例：

```json
{
  "type": "image_job_status",
  "payload": {
    "job_id": 12,
    "novel_id": 3,
    "chapter_no": 2,
    "status": "planning",
    "current_step": "build_prompt"
  }
}
```

面试映射：

- 如何设计 AI 任务状态机

## 2. `agent_trace_step`

表示执行轨迹中的一步。

面试映射：

- tracing 怎么设计

## 3. `prompt_candidate`

表示产生了一个 prompt 草稿。

面试映射：

- 为什么 prompt 要产品化可见

## 4. `checkpoint_required`

表示任务暂停等待人工确认。

面试映射：

- human-in-the-loop 怎么落地

## 5. `image_result`

表示新图片结果已产生。

面试映射：

- 结果与过程如何分开建模

## 6. `image_feedback_recorded`

表示人工反馈已记录。

面试映射：

- 反馈闭环如何进入系统

## 七、建议的最小 workflow 结构

当前最适合的链路不是复杂多 agent，而是以下可解释流程：

1. 读取 chapter 内容
2. `story_analyzer` 生成结构化摘要
3. `image_prompt_generator` 生成一个或多个 prompt candidate
4. 如需人工确认，则停在 checkpoint
5. `generate_image` tool 调用 provider 生成图片
6. `result_reviewer` 生成结果摘要或一致性检查
7. 保存 image result
8. 等待人工评分

## 哪一步体现 agent / workflow 边界判断

### 应该使用 agent 的步骤

- `story_analyzer`
- `image_prompt_generator`
- `result_reviewer`

原因：

- 需要理解开放文本
- 需要抽取结构
- 需要容忍语义层不确定性

### 应该使用 deterministic workflow 的步骤

- chapter 校验
- task/job 创建
- checkpoint 状态变更
- provider 调用
- 图片结果持久化
- feedback 写入

原因：

- 步骤边界稳定
- 输入输出清晰
- 更需要可控性而不是创造性

### 面试中可以直接回答的问题

- 为什么某些步骤不用 agent？
- tool 和 workflow 的边界怎么定？

## 八、建议新增的 tool 设计

## 1. `analyze_story`

输入：

- `chapter_text`
- `novel_summary`
- `character_sheets`

输出：

- `scene_summary`
- `character_refs`
- `visual_focus`
- `risk_flags`

面试映射：

- structured output
- tool schema 设计

## 2. `build_image_prompt`

输入：

- `scene_summary`
- `character_refs`
- `style_profile`
- `image_count`

输出：

- `prompt_candidates`

面试映射：

- prompt 生成为什么不应直接耦合到 handler

## 3. `generate_image`

输入：

- `prompt_text`
- `negative_prompt`
- `provider_name`
- `model_name`

输出：

- `image_urls`
- `seed`
- `provider_metadata`

面试映射：

- 为什么 provider 调用更适合 deterministic tool，而不是 agent 自由发挥

## 4. `evaluate_image_result`

输入：

- `prompt_text`
- `image_result_summary`
- `character_refs`

输出：

- `consistency_score`
- `issues`

面试映射：

- 最小 eval 怎样开始做

## 九、前端页面与交互建议

为了兼顾产品价值和面试价值，前端不应只做一个“提交后等图片”的页面，而应尽量显式暴露 AI 系统关键节点。

建议至少补以下 UI 区域：

## 1. 文生图发起面板

用于：

- 选择章节
- 选择文本范围
- 选择图片数量
- 选择是否需要 prompt 确认

## 2. 任务状态面板

用于展示：

- job 状态
- 当前步骤
- 最近 trace

## 3. prompt 面板

用于展示：

- prompt 草稿
- 允许人工修改并确认

## 4. 图片结果面板

用于展示：

- 生成结果
- 重试入口
- 评分入口

这样前端本身就成为 AI 工程能力的展示面，而不只是结果展示面。

## 十、建议的第一阶段落地范围

为了控制复杂度，第一阶段建议只做以下最小范围：

1. `POST /novels/:id/chapters/:no/image-jobs`
2. `GET /image-jobs/:job_id`
3. `GET /image-jobs/:job_id/prompts`
4. `POST /image-jobs/:job_id/prompts/:prompt_id/confirm`
5. `POST /image-results/:id/feedback`
6. SSE 新增：`image_job_status`、`prompt_candidate`、`checkpoint_required`、`image_result`

第一阶段先不做：

- 多 agent 协作
- 独立 agent 服务化
- 复杂 resume/interrupt
- 完整视频链路

## 十一、为什么这套设计对面试有帮助

如果只做“提交文本然后生成图片”，你在面试里大概率只能讲：

- 我做了一个文生图功能

但如果按本文档设计，你可以讲成以下问题的真实实践：

1. 领域 API 和通用 agent API 如何分层
2. 为什么某些步骤该用 agent，某些步骤该用 workflow
3. tool schema 为什么必须结构化
4. prompt 为什么要版本化和产品化
5. 为什么需要 human-in-the-loop
6. tracing 和 feedback 怎样从一开始就进入系统

这会让你的项目从“AI 功能项目”变成“AI 工程项目”。

## 十二、一句话结论

当前 `auto_drama` 的文生图改进，不应该只设计成“新增一个生成图片接口”，而应该设计成：

**一个以 `ImageJob` 为核心、通过领域接口驱动、显式暴露 agent 决策、tool 契约、prompt 版本、trace 和人工反馈的最小 AI 工程链路。**
