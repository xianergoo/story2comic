# Frontend Streaming Experience Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把小说生成主链路补齐为可完整操作、可实时观察、可进入章节查看流式正文的前端体验。

**Architecture:** 后端先统一 SSE 结构化事件协议，并补章节流式正文推送；前端围绕“首页总览、详情总控、章节阅读”三层页面重构状态消费。页面初始状态来自 API 快照，实时变化来自 SSE 增量事件，断线后通过重新拉取快照校准。

**Tech Stack:** Go, Gin, GORM, SQLite, Vue 3, Vite, Axios, EventSource

---

## File Map

### Backend

- Modify: `internal/handler/sse.go`
  - 从单字符串 channel 升级为结构化 SSE 事件推送。
- Create: `internal/task/events.go`
  - 定义统一 SSE 事件结构、事件类型、构造辅助函数。
- Modify: `internal/service/novel_service.go`
  - 推送 `novel_status`、`progress_summary`、`log`、`outline_stream` 等作品级事件。
- Modify: `internal/service/chapter_service.go`
  - 增加章节流式写作接口，支持 `chapter_stream` 推送。
- Modify: `internal/service/task_handler.go`
  - 在章节/漫画任务关键节点推送 `chapter_status`、`comic_status`、`error`、`log`。
- Modify: `internal/task/handler.go`
  - 扩展事件发布接口定义，避免服务层手拼字符串。
- Modify: `internal/app/tasks.go`
  - 注入新的事件发布逻辑，替代 `onProgress` 字符串封装。
- Modify: `internal/handler/novel.go`
  - 详情接口返回更适合前端总控页的数据结构。
- Modify: `internal/handler/chapter.go`
  - 章节接口补足占位章节态、漫画页展示结构和边界返回。

### Frontend

- Modify: `web/src/api/index.js`
  - 定义结构化 SSE 事件解析与新页面 API 调用。
- Create: `web/src/composables/useNovelStream.js`
  - 管理 SSE 连接、事件分发、连接状态与快照校准钩子。
- Modify: `web/src/router/index.js`
  - 新增章节页路由 `/novel/:id/chapter/:no`。
- Modify: `web/src/views/HomeView.vue`
  - 增强书架页状态卡、快速继续/停止按钮、空状态。
- Modify: `web/src/views/NovelDetail.vue`
  - 重做为总控页：状态卡、进度区、章节列表、事件流、CTA。
- Create: `web/src/views/ChapterView.vue`
  - 新增章节阅读页：正文、流式正文、漫画区、导航、错误和空态。
- Create: `web/src/components/EventLogPanel.vue`
  - 可折叠事件流面板。
- Create: `web/src/components/OutlineSummaryPanel.vue`
  - 展示大纲生成中摘要、完成态和空态。
- Create: `web/src/components/ChapterList.vue`
  - 章节列表及状态、高亮、阅读/重生成操作。
- Create: `web/src/components/StatusOverview.vue`
  - 详情页顶部状态卡和进度概览。
- Create: `web/src/components/StreamTextPanel.vue`
  - 章节页正文流式展示面板。
- Create: `web/src/components/ComicGallery.vue`
  - 章节页漫画显示组件。

### Verification

- Run: `go build ./...`
- Run: `npm run build`

## Task 1: Define Structured SSE Events

**Files:**
- Create: `internal/task/events.go`
- Modify: `internal/task/handler.go`
- Modify: `internal/handler/sse.go`

- [ ] **Step 1: Add event type definitions**

Create a central event definition file with typed payload helpers for:
- `novel_status`
- `progress_summary`
- `chapter_status`
- `chapter_stream`
- `outline_stream`
- `comic_status`
- `log`
- `error`

- [ ] **Step 2: Extend publisher interface**

Update `internal/task/handler.go` so the publisher abstraction can publish structured events instead of raw strings.

- [ ] **Step 3: Update SSE handler storage model**

Refactor `internal/handler/sse.go` so connections publish serialized JSON events consistently, while preserving the existing `progress` event channel name unless a stronger reason emerges.

- [ ] **Step 4: Lock event transport and dispatch rule**

Document and implement one rule end-to-end:
- transport layer may keep a fixed SSE event name for compatibility
- business dispatch must use JSON `type`
- frontend state updates must never use `log` text as a state source

- [ ] **Step 5: Run backend build verification**

Run: `go build ./...`
Expected: build passes

## Task 2: Push Novel-Level Progress Events

**Files:**
- Modify: `internal/service/novel_service.go`
- Modify: `internal/app/tasks.go`

- [ ] **Step 1: Add helper methods for novel-level events**

In `internal/service/novel_service.go`, add focused helpers to publish:
- novel status
- progress summary
- log events
- outline stream events

- [ ] **Step 2: Replace string-only progress pushes**

Stop relying on `detail` string as the only signal. Keep readable log messages, but also emit structured status and summary events.

- [ ] **Step 3: Emit events at all major lifecycle points**

Cover at least:
- generation start
- outline start / finish / fail
- resume requested
- stop requested
- all chapters queued
- generation completed

- [ ] **Step 4: Make event truth-sources explicit in code paths**

Enforce the spec's source-of-truth rules:
- `novel_status` only updates novel-level state
- `progress_summary` only carries progress counters and current chapter number
- `log` is display-only
- `error` is display plus possible snapshot re-sync trigger

- [ ] **Step 5: Verify backend build**

Run: `go build ./...`
Expected: build passes

## Task 3: Add Chapter Streaming and Chapter-Level Events

**Files:**
- Modify: `internal/service/chapter_service.go`
- Modify: `internal/service/task_handler.go`

- [ ] **Step 1: Introduce chapter streaming write path**

Add a streaming-capable chapter generation method that accepts an `onChunk` callback and emits incremental正文内容.

- [ ] **Step 2: Keep non-streaming compatibility via wrapper**

Preserve `Write(...)` as a thin wrapper around the streaming implementation only if it materially reduces call-site churn; do not introduce parallel behavior branches.

- [ ] **Step 3: Emit chapter lifecycle events**

Update `internal/service/task_handler.go` to publish:
- `chapter_status` when queued / writing / done / failed
- `chapter_stream` while text arrives
- `comic_status` during image generation
- `error` and `log` for failures and milestones

- [ ] **Step 4: Define old-content behavior for regenerate**

Implement the backend behavior required by the spec:
- old chapter content may remain readable until new content finishes
- new generation should still update current chapter state to writing/regenerating-compatible UI semantics

- [ ] **Step 5: Preserve partial-content failure semantics**

Ensure stream interruption or regenerate failure can leave previously readable content intact while still surfacing failure state to the frontend.

- [ ] **Step 6: Verify backend build**

Run: `go build ./...`
Expected: build passes

## Task 4: Improve Novel and Chapter API Snapshots

**Files:**
- Modify: `internal/handler/novel.go`
- Modify: `internal/handler/chapter.go`

- [ ] **Step 1: Expand novel detail snapshot for total-control UI**

Ensure novel detail response includes data needed for:
- current progress summary
- current chapter focus
- chapter readability and status
- page completion summary

- [ ] **Step 2: Support planned-but-not-generated chapter view state**

Update chapter handler behavior so `/api/v1/novels/:id/chapters/:no` can distinguish:
- existing generated chapter
- planned but not yet generated chapter
- invalid chapter number

- [ ] **Step 3: Distinguish novel-not-found from chapter-not-found**

Make chapter page APIs return enough information for the frontend to show separate states for:
- novel does not exist
- chapter number is outside plan
- chapter is planned but not generated yet

- [ ] **Step 4: Return chapter page-friendly payload**

Ensure chapter response is optimized for chapter page rendering:
- chapter meta
- final content if present
- stream-capable placeholder state
- prev/next chapter info
- parsed comic page image URL arrays

- [ ] **Step 5: Verify backend build**

Run: `go build ./...`
Expected: build passes

## Task 5: Add Frontend SSE Composable

**Files:**
- Modify: `web/src/api/index.js`
- Create: `web/src/composables/useNovelStream.js`

- [ ] **Step 1: Normalize SSE event payload parsing**

Update the API layer so SSE payloads are parsed into typed frontend objects instead of ad-hoc `progressMsg` strings.

- [ ] **Step 2: Create `useNovelStream` composable**

The composable should manage:
- opening/closing EventSource
- connection state
- event dispatch callbacks
- error callback
- optional reconnect snapshot callback

- [ ] **Step 3: Encode event truth-source rules in composable contract**

Make the composable API explicit about:
- JSON `type` is the only dispatch key
- `log` is display-only
- pages own their state mapping, but must respect truth-source responsibilities

- [ ] **Step 4: Keep composable narrow**

Do not move page-specific UI logic into the composable. It should only manage stream transport and event fan-out.

- [ ] **Step 5: Verify frontend build**

Run: `npm run build`
Expected: build passes

## Task 6: Rebuild the Novel Detail Page

**Files:**
- Modify: `web/src/views/NovelDetail.vue`
- Create: `web/src/components/EventLogPanel.vue`
- Create: `web/src/components/OutlineSummaryPanel.vue`
- Create: `web/src/components/ChapterList.vue`
- Create: `web/src/components/StatusOverview.vue`

- [ ] **Step 1: Add summary UI blocks**

Render:
- SSE connection state
- total novel state
- progress overview
- current chapter CTA

- [ ] **Step 2: Add outline summary area**

Render `outline_stream` in a bounded summary area with:
- generating state
- completed state
- no-outline-yet empty state

- [ ] **Step 3: Add chapter list with real navigation**

Each chapter row should support:
- readable state badge
- current-stream highlight
- read action
- regenerate action

- [ ] **Step 4: Add event log panel**

Implement a collapsible timeline/log panel that consumes only `log` and `error` display data.

- [ ] **Step 5: Implement resume UX loop**

After `resume` succeeds:
- show immediate feedback
- wait for stream events
- show CTA to current streaming chapter once known

- [ ] **Step 6: Implement stop UX loop**

After `stop` succeeds:
- show immediate stop feedback
- wait for `novel_status` or snapshot re-sync confirming paused/stopped state
- restore button states correctly on failure
- keep user on detail page as the source of truth for total status

- [ ] **Step 7: Add empty and error states**

Cover:
- no chapters
- no events
- stopped state
- completed state
- disconnected SSE state

- [ ] **Step 8: Add current-chapter guidance states**

Implement the explicit guidance from the spec:
- after resume, show where to watch total progress
- when current chapter is known, show direct CTA to chapter page
- highlight the live chapter row

- [ ] **Step 9: Verify frontend build**

Run: `npm run build`
Expected: build passes

## Task 7: Add the Chapter Reading and Streaming Page

**Files:**
- Modify: `web/src/router/index.js`
- Create: `web/src/views/ChapterView.vue`
- Create: `web/src/components/StreamTextPanel.vue`
- Create: `web/src/components/ComicGallery.vue`

- [ ] **Step 1: Add chapter route**

Create `/novel/:id/chapter/:no` route.

- [ ] **Step 2: Render final content, placeholder, or streaming state**

Support these distinct modes:
- completed chapter
- planned but not generated
- invalid/not found
- generating with stream buffer

- [ ] **Step 3: Consume chapter-specific SSE events**

Only apply `chapter_stream`, `chapter_status`, `comic_status`, and relevant `error` events for the active chapter.

- [ ] **Step 4: Implement reconnect and interruption behavior**

Handle:
- stream interrupted but partial content exists
- reconnect triggers chapter snapshot re-sync
- manual refresh remains a valid recovery path if reconnect fails

- [ ] **Step 5: Add regenerate UX**

Show:
- old content marker when applicable
- current streaming buffer
- failure fallback if regenerate fails

- [ ] **Step 6: Add prev/next navigation rules**

Handle:
- first chapter
- last planned chapter
- next planned but not yet generated chapter

- [ ] **Step 7: Add not-found and placeholder page states**

Render separate UI for:
- novel not found
- chapter not in plan
- chapter planned but waiting for generation

- [ ] **Step 8: Verify frontend build**

Run: `npm run build`
Expected: build passes

## Task 8: Enhance the Home Page

**Files:**
- Modify: `web/src/views/HomeView.vue`

- [ ] **Step 1: Decide and implement home-page stream strategy**

Use structured SSE updates for bookshelf cards if feasible within the current app shape. If not, explicitly implement fast snapshot revalidation after quick actions so the home page still satisfies the spec's state clarity.

- [ ] **Step 2: Expand novel cards**

Show:
- mode
- total state
- text progress
- comic progress
- actionable buttons

- [ ] **Step 3: Add quick stop/resume controls**

Implement clear loading/disabled behavior for card-level controls.

- [ ] **Step 4: Implement quick-action guidance loop**

After quick resume or stop:
- show success/failure feedback
- refresh or update card state from structured truth sources
- for quick resume, guide the user into detail page instead of treating home page as chapter-reading surface

- [ ] **Step 5: Add empty state improvements**

Make the empty bookshelf state clearer and action-oriented.

- [ ] **Step 6: Verify frontend build**

Run: `npm run build`
Expected: build passes

## Task 9: End-to-End Verification

**Files:**
- Modify: any touched files above if verification exposes gaps

- [ ] **Step 1: Run backend build**

Run: `go build ./...`
Expected: build passes

- [ ] **Step 2: Run frontend build**

Run: `npm run build`
Expected: build passes

- [ ] **Step 3: Manual flow verification checklist**

Verify at minimum:
- create novel → land on detail page
- home page shows mode/state/progress clearly
- home page quick resume/stop works and refreshes state
- detail page shows status and chapter list
- detail page stop loop reaches paused/stopped state
- detail page outline summary reacts to `outline_stream`
- resume updates detail page state
- current chapter CTA appears
- current chapter row highlights correctly
- chapter page can show placeholder or final content
- chapter page can show novel-not-found and chapter-not-in-plan states
- chapter page can receive streaming updates
- chapter page preserves partial content on interruption and re-syncs after reconnect
- regenerate keeps old content readable and shows new activity
- comic progress can update on chapter page
- error and empty states render in all three pages

- [ ] **Step 4: Update plan or implementation notes if behavior changed**

If any verified behavior differs from the design, update the relevant design/plan docs before handoff.
