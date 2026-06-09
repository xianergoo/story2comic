# 三项目对比与整合建议

> 日期：2026-06-08 | 状态：已落盘 | 关联文档：
> - `docs/superpowers/analysis/2026-06-08-auto-drama-analysis.md`
> - `docs/superpowers/analysis/2026-06-08-local-huobao-split-analysis.md`
> - `docs/superpowers/analysis/2026-06-08-open-source-huobao-analysis.md`

## 一、目的

本文档用于把以下三个体系放到同一分析框架下：

1. 当前主仓 `auto_drama`
2. 本地拆分版 huobao
3. 开源版 `huobao-drama`

目标不是给出“谁更好”的抽象结论，而是回答四个更实际的问题：

1. 三者各自擅长什么
2. 三者各自缺什么
3. 哪些能力值得保留
4. 如果后续继续在 `auto_drama` 上演进，最合理的吸收路径是什么

## 二、整体判断

三者不是同一种系统，而是落在了三种不同阶段。

### 1. `auto_drama`

它更像：

- AI 叙事生成底座
- 小说/章节驱动的内容生成原型
- 强实时反馈体验的轻量系统

### 2. 本地拆分版 huobao

它更像：

- 短剧生产平台骨架
- 已形成工作台和中台模型的生产系统雏形
- 正在把 Agent 从 backend 中外移的过渡态架构

### 3. 开源版 `huobao-drama`

它更像：

- 产品完成度较高的单仓平台
- 面向开源分发和快速部署的一体化短剧生产系统
- 多模态和 Agent 平台化意识较强的产品实现

因此，这三者的核心强项分别是：

- `auto_drama` 强在生成体验
- 本地拆分版 huobao 强在生产骨架
- 开源版 huobao 强在产品化整合

## 三、统一对比矩阵

### 1. 产品定位

`auto_drama`：

- 偏上游故事与章节生成
- 偏 AI 内容创作原型底座

本地拆分版 huobao：

- 偏短剧生产工作台
- 偏真实生产流程支撑

开源版 huobao：

- 偏开源产品化短剧平台
- 偏一体化交付

### 2. 核心业务模型

`auto_drama`：

- `Novel`
- `Outline`
- `Chapter`
- `ComicPage`

本地拆分版 huobao：

- `Drama`
- `Episode`
- `Character`
- `Scene`
- `Storyboard`
- `Prop`
- `Asset`

开源版 huobao：

- `drama`
- `episode`
- `scene`
- `storyboard`
- `media`

判断：

- `auto_drama` 的对象更偏叙事单元
- huobao 两套的对象更偏视频生产单元

### 3. 后端形态

`auto_drama`：

- Go API
- Gin
- SQLite
- 轻量 service + queue

本地拆分版 huobao：

- Go backend 中台
- Gin + GORM + gRPC
- SQLite / MySQL
- 生成服务 + 资产体系 + 外移中的 agent

开源版 huobao：

- TypeScript backend
- Hono + Drizzle + better-sqlite3
- 单服务一体化装配

判断：

- `auto_drama` 最轻
- 本地拆分版 huobao 最厚
- 开源版 huobao 最统一

### 4. Agent 形态

`auto_drama`：

- `AgentTask` / `Checkpoint` 已有
- 但仍是外壳能力
- 未真正驱动主业务流程

本地拆分版 huobao：

- 独立 gRPC agent 方向明确
- supervisor + tools + checkpoint + resume 已出现
- 但和 backend 的边界尚未收口

开源版 huobao：

- backend 内嵌 agent
- Agent 类型拆分明确
- prompt/model/skills 支持配置化组合

判断：

- `auto_drama` 是“预埋阶段”
- 本地拆分版 huobao 是“架构拆分阶段”
- 开源版 huobao 是“产品内嵌阶段”

### 5. 实时能力

`auto_drama`：

- SSE 协议已结构化
- 作品、章节、正文流、漫画流都有事件语义
- 前端已形成快照 + 实时事件结合方式

本地拆分版 huobao：

- 有任务与 agent 流式机制
- 但主亮点不在实时文本流体验，而在工作流组织

开源版 huobao：

- 有 agent 与生产流程能力
- 但当前观察重点不在像 `auto_drama` 这样突出的章节流式阅读体验

判断：

- 三者里，`auto_drama` 的实时交互体验最成型

### 6. 前端形态

`auto_drama`：

- 书架页
- 作品详情页
- 章节页
- 偏阅读与生成观察

本地拆分版 huobao：

- 项目工作台
- workflow 页
- storyboard 编辑
- 专业编辑器 / timeline
- 偏生产操作

开源版 huobao：

- studio/workspace 思维
- 单集工作流
- Agent 交互作为独立能力

判断：

- `auto_drama` 更像内容阅读型前端
- huobao 两套更像生产工作台前端

### 7. 多模态能力

`auto_drama`：

- 文本生成
- 漫画/插画生成
- 尚未进入配音/视频/合成主流程

本地拆分版 huobao：

- 图片
- 视频
- 音频抽取
- 合成
- 资产管理

开源版 huobao：

- 图片
- 视频
- TTS
- 合成
- 多 provider adapter

判断：

- huobao 两套都已跨入多模态生产平台阶段
- `auto_drama` 仍停留在文本 + 图片扩展阶段

### 8. 部署与工程组织

`auto_drama`：

- 最适合单机原型迭代
- 理解成本低
- 改造成本也最低

本地拆分版 huobao：

- 结构最复杂
- 能力最厚
- 当前也最容易出现边界治理成本

开源版 huobao：

- 单仓交付体验最好
- 更适合开源分发、Demo、Docker 化运行

## 四、三者各自最值得保留的核心价值

### 1. `auto_drama` 最值得保留的部分

- 结构化 SSE 协议
- 流式正文与状态反馈体验
- 轻量异步生成闭环
- 快照 + 实时事件结合方式
- 低成本单机迭代能力

### 2. 本地拆分版 huobao 最值得保留的部分

- 短剧领域业务模型
- 资产沉淀与生成记录体系
- 分阶段生产工作台
- 人机协同而不是黑盒全自动
- backend / agent 分层方向

### 3. 开源版 huobao 最值得保留的部分

- 多模态 adapter registry 思路
- FFmpeg 合成作为正式能力
- Agent 类型拆分方式
- Agent config + skills 机制
- 单仓产品化组织经验

## 五、不建议直接照搬的部分

### 1. 不建议把 `auto_drama` 直接当作最终平台形态

原因：

- 业务模型太偏小说
- 资产体系不足
- 任务系统仍偏原型

### 2. 不建议原样照搬本地拆分版 huobao 当前三仓状态

原因：

- 仍处于 Agent 外移中的过渡态
- 新旧接口并存
- 边界尚未完全收口

### 3. 不建议直接把开源版 huobao 的 TypeScript 单仓整体替换进来

原因：

- 技术栈切换成本高
- 它适合产品化开源交付，不一定适合作为你当前仓库的直接演进基座
- 内嵌 agent 结构不一定符合后续独立编排方向

## 六、整合判断：后续应以什么为主线

如果后续目标是继续在当前主仓上推进，我建议主线非常明确：

**以 `auto_drama` 为演进底座，但逐步吸收 huobao 体系的业务骨架与平台化能力。**

原因如下：

### 1. `auto_drama` 当前最易继续演进

- 代码结构刚完成一轮整理
- 实时生成体验已经有明显优势
- 当前主仓上下文和修复节奏都在这里

### 2. 本地拆分版 huobao 更适合作为业务模型和工作台参考源

它提供的是：

- 短剧实体模型
- 生产流程设计
- 工作台组织方式

这些是当前 `auto_drama` 最缺但最值得补的部分。

### 3. 开源版 huobao 更适合作为产品化能力参考源

它提供的是：

- 多模态 provider 抽象
- Agent 类型拆分思路
- skills 化扩展
- 合成能力正式建模

## 七、建议的演进路径

### 阶段 1：先把 `auto_drama` 从“小说系统”提升为“项目系统”

目标：

- 仍保留当前 SSE 和生成体验优势
- 但开始引入短剧领域对象

建议优先增加的对象层：

1. `Project` / `DramaProject`
2. `Episode`
3. `Scene`
4. `StoryboardFrame`
5. `Asset`

这一阶段重点不是一下子废弃 `Novel/Chapter`，而是建立新的映射关系与演化路径。

### 阶段 2：把当前任务系统升级为平台任务系统

目标：

- 从内存 queue 演进到可恢复任务系统
- 支撑更长链路、更复杂媒体任务

建议能力：

1. 任务表持久化
2. 明确任务状态机
3. 取消 / 重试 / 超时
4. 子任务链路或 pipeline
5. 与 SSE 统一事件模型打通

### 阶段 3：把 Agent 从“预埋能力”升级成“生产编排器”

建议不要直接复制本地拆分版 huobao 当前实现，而是吸收其方向，并结合开源版 huobao 的类型拆分思路。

建议的 agent 演进路径：

1. 剧情策划 agent
2. 角色/设定抽取 agent
3. 分镜拆解 agent
4. 画面提示词 agent
5. 审核/一致性 agent

同时保留 checkpoint 作为未来人工审阅节点。

### 阶段 4：把前端从“阅读页”提升为“生产工作台”

当前前端不必推倒重来，但需要逐步增加：

1. 项目总控台
2. Episode 工作流页
3. Storyboard 编辑区
4. 资产列表与预览区
5. 任务面板
6. Agent checkpoint 面板

建议做法是：

- 保留现有首页 / 详情 / 章节页
- 在此基础上逐步补工作台页，而不是一次性重写前端

### 阶段 5：引入多模态与合成正式能力

这一阶段最值得参考开源版 huobao：

1. provider registry
2. TTS 抽象
3. 视频生成抽象
4. FFmpeg 合成流程
5. 成片资产归档

## 八、优先级建议

如果要控制风险和节奏，我建议优先顺序如下：

### 优先级最高

1. 从 `Novel/Chapter` 视角补出 `Project/Episode/Scene` 的新业务骨架
2. 统一任务状态与事件模型
3. 为前端增加工作台入口，而不破坏现有阅读链路

### 第二优先级

1. 引入 `Storyboard` / `Asset` 模型
2. 规范多模态 provider 抽象
3. 把 checkpoint 从后端表结构变成真实流程节点

### 第三优先级

1. 再考虑 agent 独立服务化
2. 再考虑更重的分布式任务架构
3. 再考虑更完整的视频生产和成片流程

## 九、最终建议

最合理的整合策略不是三选一，而是：

### 1. 以 `auto_drama` 为当前演进主仓

保留：

- 当前代码上下文
- 实时体验优势
- 低改造成本

### 2. 从本地拆分版 huobao 吸收“短剧生产骨架”

重点吸收：

- 领域模型
- 资产系统
- 工作台流程
- 人机协同范式

### 3. 从开源版 huobao 吸收“平台化实现经验”

重点吸收：

- 多模态 adapter 模式
- Agent 类型拆分
- skills 扩展思路
- 合成能力产品化方式

## 十、一句话结论

三者并不是替代关系，而是互补关系：

**`auto_drama` 提供当前最好的实时生成底座，本地拆分版 huobao 提供最有价值的短剧生产骨架，开源版 huobao 提供最完整的平台化实现参考。后续最合理的路线，是在 `auto_drama` 上吸收后两者的结构与能力，而不是整体迁移到其中任意一个。**
