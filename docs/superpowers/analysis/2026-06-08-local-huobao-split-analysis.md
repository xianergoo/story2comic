# 本地拆分版 huobao 独立分析

> 日期：2026-06-08 | 状态：已落盘 | 范围：`huobao-drama-backend` / `huobao-drama-agent` / `huobao-drama-web`

## 一、整体判断

这套本地拆分版 huobao 已经形成了三仓结构，但从代码和接口边界看，更像是：

**已能运行的拆分中间态，而不是边界彻底稳定的成熟三服务体系。**

涉及目录：

1. `/Users/z3/workspace/explore/splay/huobao-drama-backend`
2. `/Users/z3/workspace/explore/splay/huobao-drama-agent`
3. `/Users/z3/workspace/explore/splay/huobao-drama-web`

## 二、三个仓库的职责定位

### 1. backend：生产主后端 / 数据与流程中台

backend 是整个平台当前真正的核心枢纽，主要负责：

1. 短剧项目管理
2. 剧集与分镜相关业务对象管理
3. 生成任务管理
4. 资产沉淀与资源管理
5. AI 配置管理
6. 对前端提供完整工作流 API
7. 对 agent 提供接入位

它承担的是一套标准“短剧生产中台”的职责，而不是一个简单接口层。

### 2. agent：独立 AI 编排执行器

agent 仓库的目标定位比较明确：

1. 作为独立 gRPC Agent 服务运行
2. 管理 supervisor / tools 编排
3. 处理流式任务事件
4. 支持 checkpoint / resume
5. 通过 gRPC 调用 backend 的数据服务

它的定位不是数据库中心，而是：

**智能任务执行中枢。**

### 3. web：人工工作台 / 生产操作台

web 明确不是普通管理后台，而是围绕短剧生产过程构建的工作台，主要负责：

1. 项目创建与管理
2. 章节/角色/场景/分镜编辑
3. 图像与视频生成操作
4. 专业编辑与时间线编辑
5. Agent 任务操作与状态查看
6. AI 设置页

## 三、三仓拆分是否完整

当前判断：

**拆分方向正确，但拆分收口还没有完成。**

主要证据如下。

### 1. backend 内仍保留旧 Agent 路径

backend 中仍存在本地 agent/orchestrator 相关逻辑，说明独立 agent 尚未彻底取代内置实现。

### 2. backend 仍带有本地 agent 代码

`pkg/agent/*` 仍然存在，包含：

- planner
- executor
- orchestrator
- checkpoint manager
- reflector
- tools

说明“agent 完整外移”尚未完成。

### 3. backend 侧 gRPC 数据服务尚未真正闭环

agent 理论上依赖 backend 的 `DramaDataService` 独立运行，但相关实现里仍有 `not implemented`，这意味着协议层边界尚未收口。

### 4. web 对 Agent 接口仍有新旧混用痕迹

前端 Agent API 中同时存在新接口和旧接口假设，说明前端尚未完全切换到统一的新 agent 协议。

### 5. agent 仓库 README 与当前结构存在不一致

这进一步表明它处于持续重构中的阶段，而不是已经稳定下来的成熟独立服务。

## 四、backend 项目画像

### 1. 核心业务模型已经比较完整

backend 的领域对象明显围绕真实短剧生产链路组织，核心包括：

- `Drama`
- `Episode`
- `Character`
- `Scene`
- `Storyboard`
- `Prop`
- `Asset`
- 图像/视频/合成记录
- `AgentTask`
- `AgentCheckpoint`

相比 `auto_drama`，它已经不是“章节写作系统”，而是：

**围绕短剧生产流程的完整业务骨架。**

### 2. 技术栈特征

backend 的主要技术栈包括：

- Go
- Gin
- GORM
- SQLite / MySQL
- gRPC + protobuf
- Viper
- 本地文件存储
- ffmpeg 外部调用
- 多 AI provider 适配层

从工程风格看，它明显比 `auto_drama` 更厚重，也更接近平台中台。

### 3. 基础设施成熟度

相对成熟的部分：

- 配置体系
- 数据库初始化与自动迁移
- 存储组件封装
- 静态资源与 API 同服能力
- 路由和 service 基本分层

仍偏过渡的部分：

- 服务边界未完全稳定
- gRPC 数据服务未闭环
- 仍偏单体主仓驱动
- 更像快速演进期工程，而非完全稳定平台

## 五、agent 项目的独立价值与成熟度

### 1. 独立价值

agent 项目真正值钱的地方，不是“多了一个服务”，而是它在尝试把：

- 规划
- 执行
- 中断
- 恢复
- checkpoint
- 流式任务事件

从主业务后端中分离出来。

这意味着它试图把 AI 能力从“某个 handler 里调用模型”，推进为：

**独立的生产编排层。**

### 2. 成熟度判断

有价值的部分：

- gRPC 服务形态明确
- 任务生命周期模型完整
- supervisor + tools 结构清楚
- 引入 Redis 做 checkpoint store

明显不足的部分：

- 仍依赖 backend 未完成的数据服务
- 任务状态存储偏轻
- cancel 语义更像状态取消，而不是强执行中断
- README 与入口结构不一致
- 更像骨架，而不是完全成熟生产 agent

### 3. 与 backend 的边界关系

理论上应当是：

- backend：业务事实源 + 资产系统 + 生成能力
- agent：编排器，只调 backend 暴露的数据和能力

但现实中仍然存在：

- backend 保留本地 agent
- agent 不能完全独立闭环
- web 同时兼容两代接口

因此当前边界是：

**方向正确，落地未收口。**

## 六、web 项目的工作台价值

web 是这套拆分版 huobao 非常重要的一部分，它提供的不是普通 CRUD，而是：

**完整的人机协同短剧生产工作台。**

它覆盖的能力包括：

1. 项目管理
2. 剧集制作流程
3. 角色/场景/道具管理
4. 分镜编辑
5. 图像与视频生成操作
6. 专业编辑器 / 时间线编辑器
7. Agent 任务查看与操作

页面组织方式明显围绕业务阶段展开，例如：

- drama
- workflow
- script
- storyboard
- generation
- editor
- settings
- agent

这说明它的设计思路是：

**项目驱动 + 生产阶段驱动。**

## 七、核心优点

### 1. 业务骨架完整

它已经围绕真实短剧生产建立了一整套对象链路：

项目 -> 集 -> 角色 / 场景 / 分镜 -> 图 -> 视频 -> 资产 -> 合成

### 2. 强调工作流和人工可控

它不是黑盒一键生成，而是允许：

- 分阶段推进
- 中间编辑
- 人工审核
- 专业调整

### 3. backend 已形成生成中台雏形

AI provider、生成服务、媒体处理、资产沉淀等关键中台能力已经出现。

### 4. agent 外移方向正确

虽然未收尾，但“把智能编排从主后端拆出去”本身是合理方向。

### 5. 平台化意识明显

从 AI 配置、资产、任务、角色库到时间线编辑，都能看出它的目标不是 demo，而是平台。

## 八、明显局限与中间态问题

### 1. 新旧 Agent 双轨并存

这是当前最明显的中间态特征，会带来：

- backend 负担重
- web 兼容成本高
- 系统边界不清晰

### 2. 服务拆分没有真正闭环

独立 agent 依赖 backend 的 gRPC 数据服务，但后者尚未完成，说明拆分仍停留在结构层，而不是稳定运行边界层。

### 3. agent 任务系统仍偏轻

当前更像内存态执行器 + Redis checkpoint，而不是成熟任务平台。

### 4. 前端 Agent 工作台尚未统一协议

说明新体系还未彻底收口。

### 5. 工程完成度不均衡

backend 很厚，web 很丰富，但 agent 明显仍在骨架期。

## 九、从 AI 短剧生成平台视角的核心价值

本地拆分版 huobao 最值得保留的，不是某个模型接入细节，而是以下两层核心资产。

### 1. 短剧生产流程的数据与业务骨架

也就是：

- `Drama`
- `Episode`
- `Character`
- `Scene`
- `Storyboard`
- `Prop`
- `Asset`

这是最难得也最有迁移价值的部分。

### 2. 人机协同工作台范式

它体现出的不是“自动生成一切”，而是：

- AI 先产出
- 人工可接管
- 中间产物可编辑
- 最终成片可专业调整

这比一键黑盒生成更接近真实生产系统。

## 十、后续适合迁移/复用的能力类别

### 1. 优先复用的业务模型层

- 项目 / 集 / 角色 / 场景 / 分镜 / 道具
- 资产模型
- 生成记录模型
- 工作流状态模型

### 2. 可复用的 backend 中台能力

- AI provider 抽象
- 图片 / 视频生成服务封装
- 本地化资源与资产归档
- 音频抽取与视频合成等媒体处理
- AI 配置管理

### 3. 可复用的前端工作台能力

- 项目管理页
- Episode workflow
- Storyboard 编辑页
- Professional editor / timeline editor
- 角色/场景/道具管理交互

### 4. 可选择性吸收的 Agent 能力

建议吸收思路，而不要直接照搬当前实现：

- supervisor + tools
- checkpoint / resume
- 流式进度事件
- 人工审批节点

## 十一、一句话结论

本地拆分版 huobao 的本质，不是“已经完成微服务化的短剧平台”，而是：

**一个以 backend 为核心、web 已较成熟、agent 正在外移中的 AI 短剧生产工作台。**
