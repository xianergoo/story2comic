<template>
  <div class="min-h-screen bg-slate-100">
    <nav class="border-b border-slate-200 bg-white/90 backdrop-blur">
      <div class="mx-auto flex max-w-7xl flex-col gap-4 px-4 py-4 lg:flex-row lg:items-start lg:justify-between">
        <div class="min-w-0">
          <router-link to="/" class="text-sm text-slate-400 transition hover:text-indigo-600">&larr; 返回作品列表</router-link>
          <h1 class="mt-2 text-3xl font-semibold tracking-tight text-slate-900">
            {{ novel?.title || '作品详情' }}
          </h1>
          <div class="mt-3 flex flex-wrap gap-2 text-sm text-slate-500">
            <span class="rounded-full bg-slate-100 px-3 py-1">模式: {{ novel?.mode || '-' }}</span>
            <span class="rounded-full bg-slate-100 px-3 py-1">总状态: {{ novelStatusLabel }}</span>
            <span class="rounded-full bg-slate-100 px-3 py-1">文字: {{ textStatusLabel }}</span>
          </div>
        </div>

        <div class="flex flex-col items-start gap-3 lg:items-end">
          <span class="rounded-full px-3 py-1 text-sm font-medium" :class="streamStatusClass">
            {{ streamStatusLabel }}
          </span>

          <div class="flex flex-wrap gap-3">
            <button
              type="button"
              class="inline-flex items-center rounded-full bg-slate-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:bg-slate-300"
              :disabled="resumeDisabled"
              @click="handleResume"
            >
              {{ actionState === 'resume' ? '继续中...' : '继续生成' }}
            </button>
            <button
              type="button"
              class="inline-flex items-center rounded-full bg-rose-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-rose-500 disabled:cursor-not-allowed disabled:bg-rose-200"
              :disabled="stopDisabled"
              @click="handleStop"
            >
              {{ actionState === 'stop' ? '停止中...' : '停止生成' }}
            </button>
          </div>
        </div>
      </div>
    </nav>

    <main class="mx-auto max-w-7xl px-4 py-6">
      <div v-if="loading" class="rounded-3xl border border-slate-200 bg-white px-6 py-20 text-center text-slate-400 shadow-sm">
        正在加载作品快照...
      </div>

      <div v-else-if="novel" class="space-y-6">
        <div v-if="feedbackMessage" class="rounded-2xl border border-indigo-200 bg-indigo-50 px-4 py-3 text-sm text-indigo-700">
          {{ feedbackMessage }}
        </div>

        <div v-if="errorMessage" class="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
          {{ errorMessage }}
        </div>

        <StatusOverview
          :progress="progress"
          :current-chapter="currentChapter"
          :cta="currentChapterCta"
          @cta="handleCurrentChapterCta"
        />

        <div class="grid gap-6 xl:grid-cols-[minmax(0,1.3fr)_minmax(320px,0.7fr)]">
          <ChapterList
            :novel-id="route.params.id"
            :chapters="chapters"
            :current-chapter-no="progress.current_chapter_no || focus?.chapter_no || 0"
            :regenerating="actionState === 'regenerate'"
            :is-busy="Boolean(actionState)"
            @regenerate="handleRegen"
          />

          <div class="space-y-6">
            <OutlineSummaryPanel
              :outline="outline"
              :stream-content="outlineDisplayContent"
              :stream-done="outlineStream.done"
            />
            <EventLogPanel :items="timeline" />
          </div>
        </div>
      </div>

      <div v-else class="rounded-3xl border border-rose-200 bg-white px-6 py-20 text-center text-rose-500 shadow-sm">
        未找到作品详情
      </div>
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue';
import { useRoute } from 'vue-router';
import ChapterList from '../components/ChapterList.vue';
import EventLogPanel from '../components/EventLogPanel.vue';
import OutlineSummaryPanel from '../components/OutlineSummaryPanel.vue';
import StatusOverview from '../components/StatusOverview.vue';
import { novelAPI } from '../api';
import { useNovelStream } from '../composables/useNovelStream';

const route = useRoute();

const novel = ref(null);
const chapters = ref([]);
const outline = ref(null);
const focus = ref(null);
const loading = ref(true);
const errorMessage = ref('');
const feedbackMessage = ref('');
const actionState = ref('');
const timeline = ref([]);
const outlineStream = reactive({
  baseContent: '',
  deltaContent: '',
  done: false,
});
const progress = reactive({
  planned_count: 0,
  generated_count: 0,
  text_done_count: 0,
  remaining_count: 0,
  current_chapter_no: 0,
});
let detailRefreshPromise = null;

const novelStatusLabel = computed(() => formatNovelStatus(novel.value?.status));
const textStatusLabel = computed(() => formatTextStatus(novel.value?.text_status));
const outlineDisplayContent = computed(() => `${outlineStream.baseContent}${outlineStream.deltaContent}`);

const stream = useNovelStream({
  novelId: route.params.id,
  eventNames: ['progress'],
  onReconnectSnapshot: loadDetail,
  onError: (error) => {
    if (stream.status.value === 'error') {
      feedbackMessage.value = feedbackMessage.value || '实时连接暂时中断，恢复后会自动回拉快照。';
    }
    if (!errorMessage.value && loading.value === false && stream.status.value === 'closed') {
      errorMessage.value = '';
    }
  },
  handlers: {
    novel_status: async (payload) => {
      reconcileActionState({
        novelStatus: payload?.status,
        textStatus: payload?.text_status,
        source: 'event',
      });
      await refreshDetailFromEvent();
    },
    progress_summary: async () => {
      feedbackMessage.value = feedbackMessage.value || '进度已刷新';
      await refreshDetailFromEvent();
    },
    chapter_status: async () => {
      await refreshDetailFromEvent();
    },
    comic_status: async () => {
      await refreshDetailFromEvent();
    },
    outline_stream: (payload) => {
      if (payload?.content) {
        outlineStream.deltaContent = `${outlineStream.deltaContent}${payload.content}`;
        outlineStream.done = false;
      }
      if (payload?.done) {
        outlineStream.done = true;
      }
      if (payload?.done) {
        feedbackMessage.value = '大纲生成完成';
      }
    },
    log: (payload) => {
      pushTimeline('log', payload);
      if (payload?.message) {
        feedbackMessage.value = payload.message;
      }
    },
    error: (payload) => {
      pushTimeline('error', payload);
      errorMessage.value = payload?.message || '任务执行失败';
      actionState.value = '';
    },
  },
});

const streamStatusLabel = computed(() => {
  if (stream.status.value === 'open') return 'SSE 已连接';
  if (stream.status.value === 'connecting') return 'SSE 连接中';
  if (stream.status.value === 'error') return 'SSE 已断开';
  return 'SSE 未连接';
});

const streamStatusClass = computed(() => {
  if (stream.status.value === 'open') return 'bg-emerald-100 text-emerald-700';
  if (stream.status.value === 'connecting') return 'bg-amber-100 text-amber-700';
  if (stream.status.value === 'error') return 'bg-rose-100 text-rose-700';
  return 'bg-slate-100 text-slate-600';
});

const currentChapter = computed(() => {
  const chapterNo = progress.current_chapter_no || focus.value?.chapter_no || 0;
  if (!chapterNo) {
    return focus.value;
  }
  return chapters.value.find((item) => item.chapter_no === chapterNo) || focus.value;
});

const currentChapterCta = computed(() => {
  if (!currentChapter.value) {
    return null;
  }
  if (currentChapter.value.has_content || currentChapter.value.state === 'generated') {
    return {
      label: '前往阅读',
      disabled: false,
    };
  }
  if (novel.value?.status === 'drafting') {
    return {
      label: '生成进行中',
      disabled: true,
    };
  }
  return {
    label: '等待生成',
    disabled: true,
  };
});

const resumeDisabled = computed(() => {
  return actionState.value !== '' || novel.value?.status === 'drafting';
});

const stopDisabled = computed(() => {
  return actionState.value !== '' || novel.value?.text_status !== 'writing';
});

onMounted(async () => {
  try {
    await loadDetail();
    stream.open();
  } catch (error) {
    errorMessage.value = extractErrorMessage(error, '加载失败');
  } finally {
    loading.value = false;
  }
});

async function loadDetail(options = {}) {
  const { silent = false } = options;
  const detail = await novelAPI.detail(route.params.id);
  const payload = detail?.data || {};

  novel.value = payload.novel || null;
  chapters.value = Array.isArray(payload.chapters) ? payload.chapters : [];
  outline.value = payload.outline || null;
  focus.value = payload.focus || null;
  applyProgressSnapshot(payload.progress);
  syncOutlineStreamFromSnapshot(payload.outline);
  reconcileActionState({
    novelStatus: payload?.novel?.status,
    textStatus: payload?.novel?.text_status,
    source: 'snapshot',
  });

  if (!silent) {
    errorMessage.value = '';
  }
}

function applyProgressSnapshot(snapshot = {}) {
  progress.planned_count = Number(snapshot?.planned_count ?? 0);
  progress.generated_count = Number(snapshot?.generated_count ?? 0);
  progress.text_done_count = Number(snapshot?.text_done_count ?? 0);
  progress.remaining_count = Number(snapshot?.remaining_count ?? 0);
  progress.current_chapter_no = Number(snapshot?.current_chapter_no ?? 0);
}

async function refreshDetailFromEvent() {
  if (!detailRefreshPromise) {
    detailRefreshPromise = loadDetail({ silent: true }).finally(() => {
      detailRefreshPromise = null;
    });
  }
  await detailRefreshPromise;
}

function syncOutlineStreamFromSnapshot(snapshotOutline) {
  const snapshotContent = snapshotOutline?.content || '';
  const previousDisplayContent = `${outlineStream.baseContent}${outlineStream.deltaContent}`;

  outlineStream.baseContent = snapshotContent;

  if (snapshotContent && previousDisplayContent.startsWith(snapshotContent)) {
    outlineStream.deltaContent = previousDisplayContent.slice(snapshotContent.length);
  } else if (snapshotContent && snapshotContent.startsWith(previousDisplayContent)) {
    outlineStream.deltaContent = '';
  } else if (!snapshotContent) {
    outlineStream.deltaContent = outlineStream.deltaContent;
  } else {
    outlineStream.deltaContent = '';
  }

  if (snapshotContent && outlineStream.deltaContent.length === 0) {
    outlineStream.done = true;
  }
}

function pushTimeline(level, payload = {}) {
  timeline.value = [
    {
      id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
      level,
      step: payload?.step || '',
      chapter_no: payload?.chapter_no || 0,
      message: payload?.message || (level === 'error' ? '未知错误' : '收到事件'),
      createdAt: new Date().toLocaleTimeString('zh-CN', { hour12: false }),
    },
    ...timeline.value,
  ].slice(0, 100);
}

function reconcileActionState({ novelStatus, textStatus, source }) {
  if (actionState.value === 'resume') {
    if (novelStatus === 'drafting') {
      feedbackMessage.value = '继续生成请求已提交';
      actionState.value = '';
      return;
    }

    if (novelStatus === 'completed' || textStatus === 'done') {
      feedbackMessage.value = source === 'snapshot' ? '作品已处于完成态' : '作品生成已完成';
      actionState.value = '';
      return;
    }
  }

  if (actionState.value === 'stop') {
    if (textStatus === 'stopped' || novelStatus === 'stopped') {
      feedbackMessage.value = '生成任务已停止';
      actionState.value = '';
      return;
    }

    if (novelStatus === 'completed' || textStatus === 'done') {
      feedbackMessage.value = source === 'snapshot' ? '作品已完成，无需停止' : '作品生成已完成';
      actionState.value = '';
    }
  }
}

async function handleStop() {
  errorMessage.value = '';
  feedbackMessage.value = '正在请求停止生成...';
  actionState.value = 'stop';
  try {
    await novelAPI.stop(route.params.id);
    await loadDetail({ silent: true });
    feedbackMessage.value = '停止请求已发送，等待任务状态同步';
  } catch (error) {
    errorMessage.value = extractErrorMessage(error, '停止失败');
    actionState.value = '';
  }
}

async function handleResume() {
  errorMessage.value = '';
  feedbackMessage.value = '正在请求继续生成...';
  actionState.value = 'resume';
  try {
    await novelAPI.resume(route.params.id);
    await loadDetail({ silent: true });
    feedbackMessage.value = '继续生成请求已发送，等待流事件确认';
  } catch (error) {
    errorMessage.value = extractErrorMessage(error, '继续生成失败');
    actionState.value = '';
  }
}

async function handleRegen(chapterNo) {
  errorMessage.value = '';
  feedbackMessage.value = `正在重生成第 ${chapterNo} 章...`;
  actionState.value = 'regenerate';
  try {
    await novelAPI.regenChapter(route.params.id, chapterNo);
    await loadDetail({ silent: true });
    feedbackMessage.value = `第 ${chapterNo} 章已重新入队`;
  } catch (error) {
    errorMessage.value = extractErrorMessage(error, '重生成失败');
  } finally {
    actionState.value = '';
  }
}

function handleCurrentChapterCta() {
  if (!currentChapter.value) {
    return;
  }
  window.location.href = `/novel/${route.params.id}/chapter/${currentChapter.value.chapter_no}`;
}

function extractErrorMessage(error, fallback) {
  return error?.message || error?.error || fallback;
}

function formatNovelStatus(status) {
  if (status === 'drafting') return '生成中';
  if (status === 'completed') return '已完成';
  if (status === 'stopped') return '已停止';
  return status || '-';
}

function formatTextStatus(status) {
  if (status === 'writing') return '写作中';
  if (status === 'done') return '已完成';
  if (status === 'stopped') return '已停止';
  return status || '-';
}
</script>
