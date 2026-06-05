<template>
  <div class="min-h-screen bg-slate-100">
    <nav class="border-b border-slate-200 bg-white/90 backdrop-blur">
      <div class="mx-auto flex max-w-7xl flex-col gap-4 px-4 py-4 lg:flex-row lg:items-start lg:justify-between">
        <div class="min-w-0">
          <router-link :to="detailLink" class="text-sm text-slate-400 transition hover:text-indigo-600">&larr; 返回作品详情</router-link>
          <h1 class="mt-2 text-3xl font-semibold tracking-tight text-slate-900">
            {{ chapterHeading }}
          </h1>
          <div class="mt-3 flex flex-wrap gap-2 text-sm text-slate-500">
            <span class="rounded-full bg-slate-100 px-3 py-1">作品: {{ novelTitle }}</span>
            <span class="rounded-full bg-slate-100 px-3 py-1">章节状态: {{ chapterStatusLabel }}</span>
            <span class="rounded-full bg-slate-100 px-3 py-1">漫画状态: {{ comicStatusLabel }}</span>
          </div>
        </div>

        <div class="flex flex-col items-start gap-3 lg:items-end">
          <span class="rounded-full px-3 py-1 text-sm font-medium" :class="streamStatusClass">
            {{ streamStatusLabel }}
          </span>

          <div class="flex flex-wrap gap-3">
            <router-link
              class="inline-flex items-center rounded-full border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 transition hover:border-slate-400 hover:text-slate-900 disabled:pointer-events-none disabled:opacity-50"
              :class="!navigation.prev ? 'pointer-events-none opacity-50' : ''"
              :to="navLink(navigation.prev)"
            >
              上一章
            </router-link>
            <router-link
              class="inline-flex items-center rounded-full border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 transition hover:border-slate-400 hover:text-slate-900 disabled:pointer-events-none disabled:opacity-50"
              :class="!navigation.next ? 'pointer-events-none opacity-50' : ''"
              :to="navLink(navigation.next)"
            >
              下一章
            </router-link>
            <button
              type="button"
              class="inline-flex items-center rounded-full bg-slate-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:bg-slate-300"
              :disabled="regenerateDisabled"
              @click="handleRegenerate"
            >
              {{ actionState === 'regenerate' ? '重生成中...' : '重生成本章' }}
            </button>
          </div>
        </div>
      </div>
    </nav>

    <main class="mx-auto max-w-7xl px-4 py-6">
      <div v-if="loading" class="rounded-3xl border border-slate-200 bg-white px-6 py-20 text-center text-slate-400 shadow-sm">
        正在加载章节快照...
      </div>

      <div v-else class="space-y-6">
        <div v-if="feedbackMessage" class="rounded-2xl border border-indigo-200 bg-indigo-50 px-4 py-3 text-sm text-indigo-700">
          {{ feedbackMessage }}
        </div>

        <div class="grid gap-6 xl:grid-cols-[minmax(0,1.35fr)_minmax(320px,0.65fr)]">
          <StreamTextPanel
            :state="effectiveState"
            :status="effectiveChapterStatus"
            :content="displayContent"
            :placeholder="chapterData.placeholder"
            :is-streaming="isStreaming"
            :interrupted="streamInterrupted"
            :message="textPanelMessage"
            :error="chapterErrorMessage"
          />

          <ComicGallery
            :pages="comicPages"
            :image-urls="comicImageUrls"
            :status="effectiveComicStatus"
            :state="effectiveState"
            :message="comicPanelMessage"
          />
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import ComicGallery from '../components/ComicGallery.vue';
import StreamTextPanel from '../components/StreamTextPanel.vue';
import { chapterAPI, novelAPI } from '../api';
import { useNovelStream } from '../composables/useNovelStream';

const route = useRoute();

const loading = ref(true);
const actionState = ref('');
const feedbackMessage = ref('');
const chapterErrorMessage = ref('');
const chapterData = ref(createEmptyChapterData());
const streamedDelta = ref('');
const chapterStatusOverride = ref('');
const comicStatusOverride = ref('');
const chapterTerminalStatusLock = ref('');
const streamInterrupted = ref(false);
let refreshPromise = null;
let refreshPromiseRouteKey = '';
let activeNovelId = '';
let activeViewVersion = 0;

const chapterNo = computed(() => Number(route.params.no || 0));
const detailLink = computed(() => `/novel/${route.params.id}`);
const navigation = computed(() => chapterData.value.navigation || {});
const novelTitle = computed(() => chapterData.value.novel?.title || '未命名作品');
const effectiveState = computed(() => {
  if (isStreaming.value) {
    return 'generated';
  }
  return chapterData.value.state || 'invalid';
});
const effectiveChapterStatus = computed(() => {
  if (chapterStatusOverride.value) {
    return chapterStatusOverride.value;
  }
  return chapterData.value.chapter?.status || '';
});
const effectiveComicStatus = computed(() => {
  if (comicStatusOverride.value) {
    return comicStatusOverride.value;
  }
  if (comicPages.value.length > 0 || comicImageUrls.value.length > 0) {
    return 'done';
  }
  if (effectiveState.value === 'invalid') {
    return 'invalid';
  }
  return 'planned';
});
const isStreaming = computed(() => {
  return effectiveChapterStatus.value === 'writing' || streamedDelta.value.length > 0;
});
const displayContent = computed(() => `${chapterData.value.content || ''}${streamedDelta.value}`);
const chapterHeading = computed(() => {
  const title = chapterData.value.chapter?.title?.trim();
  if (title) {
    return `第 ${safeChapterNo()} 章 · ${title}`;
  }
  if (safeChapterNo() > 0) {
    return `第 ${safeChapterNo()} 章`;
  }
  return '章节不可用';
});
const chapterStatusLabel = computed(() => formatChapterStatus(effectiveChapterStatus.value, effectiveState.value));
const comicStatusLabel = computed(() => formatComicStatus(effectiveComicStatus.value));
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
const comicPages = computed(() => chapterData.value.comic?.pages || []);
const comicImageUrls = computed(() => chapterData.value.comic?.image_urls || []);
const regenerateDisabled = computed(() => {
  return actionState.value !== '' || chapterNo.value <= 0 || chapterData.value.state === 'invalid';
});
const textPanelMessage = computed(() => {
  if (isStreaming.value) {
    return streamInterrupted.value
      ? '正文流曾短暂中断，当前已保留已接收内容。'
      : '正文正在流式更新中。';
  }
  if (chapterData.value.state === 'planned') {
    return '当前章节仍在等待生成或重新排队。';
  }
  return '';
});
const comicPanelMessage = computed(() => {
  if (effectiveComicStatus.value === 'generating') {
    return '漫画生成事件已到达，图片完成后会自动刷新。';
  }
  return '';
});

const stream = useNovelStream({
  novelId: route.params.id,
  eventNames: ['progress'],
  onReconnectSnapshot: async () => {
    await loadChapter({ silent: true });
  },
  onError: () => {
    if (stream.status.value === 'error') {
      streamInterrupted.value = true;
      feedbackMessage.value = '实时连接暂时中断，恢复后会自动回拉章节快照。';
    }
  },
  handlers: {
    chapter_stream: (payload) => {
      if (!isCurrentChapterPayload(payload) || typeof payload?.content !== 'string') {
        return;
      }
      streamedDelta.value = `${streamedDelta.value}${payload.content}`;
      setChapterStatusOverride('writing');
      streamInterrupted.value = false;
      chapterErrorMessage.value = '';
      feedbackMessage.value = `第 ${safeChapterNo()} 章正文更新中`;
    },
    chapter_status: async (payload) => {
      if (!isCurrentChapterPayload(payload)) {
        return;
      }
      setChapterStatusOverride(payload?.status || '');
      streamInterrupted.value = false;

      if (payload?.status === 'writing') {
        feedbackMessage.value = `第 ${safeChapterNo()} 章开始生成`;
        return;
      }

      if (payload?.status === 'done') {
        feedbackMessage.value = `第 ${safeChapterNo()} 章生成完成`;
      } else if (payload?.status === 'failed') {
        chapterErrorMessage.value = `第 ${safeChapterNo()} 章生成失败`;
      }

      await refreshFromEvent();
    },
    comic_status: async (payload) => {
      if (!isCurrentChapterPayload(payload)) {
        return;
      }
      comicStatusOverride.value = payload?.status || '';
      feedbackMessage.value = formatComicFeedback(payload?.status);
      await refreshFromEvent();
    },
    error: (payload) => {
      if (!payload?.chapter_no) {
        return;
      }

      if (!isCurrentChapterPayload(payload)) {
        return;
      }

      chapterErrorMessage.value = payload?.message || '章节处理失败';
      setChapterStatusOverride('failed');
    },
  },
});

onMounted(async () => {
  await initialize();
});

watch(
  () => [route.params.id, route.params.no],
  async ([novelId]) => {
    await initialize({ reopenStream: novelId !== activeNovelId });
  }
);

async function initialize(options = {}) {
  const { reopenStream = false } = options;
  const viewVersion = ++activeViewVersion;
  const routeKey = currentRouteKey();

  resetLocalState();
  loading.value = true;

  try {
    await loadChapter({ viewVersion, routeKey });

    if (!isActiveViewRequest(viewVersion, routeKey)) {
      return;
    }

    if (!activeNovelId || reopenStream || activeNovelId !== String(route.params.id)) {
      stream.open(route.params.id);
      activeNovelId = String(route.params.id);
    }
  } catch (error) {
    if (!isActiveViewRequest(viewVersion, routeKey)) {
      return;
    }
    chapterErrorMessage.value = extractErrorMessage(error, '加载章节失败');
  } finally {
    if (isActiveViewRequest(viewVersion, routeKey)) {
      loading.value = false;
    }
  }
}

async function loadChapter(options = {}) {
  const {
    silent = false,
    viewVersion = activeViewVersion,
    routeKey = currentRouteKey(),
    novelId = route.params.id,
    targetChapterNo = route.params.no,
  } = options;

  const response = await chapterAPI.view(novelId, targetChapterNo);

  if (!isActiveViewRequest(viewVersion, routeKey)) {
    return false;
  }

  const payload = response?.data || createEmptyChapterData();
  syncContentFromSnapshot(payload.content || '');
  chapterData.value = {
    ...createEmptyChapterData(),
    ...payload,
    chapter: {
      ...createEmptyChapterData().chapter,
      ...(payload.chapter || {}),
    },
    navigation: payload.navigation || {},
    comic: {
      ...createEmptyChapterData().comic,
      ...(payload.comic || {}),
    },
  };

  if (!silent) {
    chapterErrorMessage.value = '';
  }

  if (shouldApplySnapshotChapterStatus(chapterData.value.chapter?.status)) {
    setChapterStatusOverride(chapterData.value.chapter.status);
  }

  if (comicPages.value.length > 0 || comicImageUrls.value.length > 0) {
    comicStatusOverride.value = 'done';
  } else if (comicStatusOverride.value === 'done') {
    comicStatusOverride.value = '';
  }

  if (chapterStatusOverride.value !== 'writing' && chapterData.value.chapter?.status !== 'writing') {
    streamedDelta.value = '';
  }

  return true;
}

async function refreshFromEvent() {
  const viewVersion = activeViewVersion;
  const routeKey = currentRouteKey();

  if (!refreshPromise || refreshPromiseRouteKey !== routeKey) {
    const currentPromise = loadChapter({ silent: true, viewVersion, routeKey }).finally(() => {
      if (refreshPromise === currentPromise) {
        refreshPromise = null;
        refreshPromiseRouteKey = '';
      }
    });
    refreshPromise = currentPromise;
    refreshPromiseRouteKey = routeKey;
  }

  await refreshPromise;
}

function syncContentFromSnapshot(snapshotContent) {
  const previousDisplay = displayContent.value;

  if (!snapshotContent) {
    return;
  }

  if (previousDisplay.startsWith(snapshotContent)) {
    streamedDelta.value = previousDisplay.slice(snapshotContent.length);
    return;
  }

  if (snapshotContent.startsWith(previousDisplay)) {
    streamedDelta.value = '';
    return;
  }

  streamedDelta.value = '';
}

async function handleRegenerate() {
  if (regenerateDisabled.value) {
    return;
  }

  const viewVersion = activeViewVersion;
  const routeKey = currentRouteKey();
  const novelId = route.params.id;
  const targetChapterNo = route.params.no;
  const targetChapterLabel = safeChapterNo();

  actionState.value = 'regenerate';
  chapterErrorMessage.value = '';
  feedbackMessage.value = `正在重生成第 ${targetChapterLabel} 章...`;

  try {
    await novelAPI.regenChapter(novelId, targetChapterNo);

    if (!isActiveViewRequest(viewVersion, routeKey)) {
      return;
    }

    setChapterStatusOverride('writing');
    comicStatusOverride.value = 'planned';
    await loadChapter({
      silent: true,
      viewVersion,
      routeKey,
      novelId,
      targetChapterNo,
    });

    if (!isActiveViewRequest(viewVersion, routeKey)) {
      return;
    }

    feedbackMessage.value = `第 ${targetChapterLabel} 章已重新入队，等待流式正文与漫画更新`;
  } catch (error) {
    if (!isActiveViewRequest(viewVersion, routeKey)) {
      return;
    }

    chapterErrorMessage.value = extractErrorMessage(error, '重生成失败');
  } finally {
    if (isActiveViewRequest(viewVersion, routeKey)) {
      actionState.value = '';
    }
  }
}

function navLink(target) {
  if (!target?.chapter_no) {
    return route.fullPath;
  }
  return `/novel/${route.params.id}/chapter/${target.chapter_no}`;
}

function isCurrentChapterPayload(payload) {
  return Number(payload?.chapter_no) === chapterNo.value;
}

function resetLocalState() {
  chapterData.value = createEmptyChapterData();
  streamedDelta.value = '';
  chapterStatusOverride.value = '';
  chapterTerminalStatusLock.value = '';
  comicStatusOverride.value = '';
  streamInterrupted.value = false;
  chapterErrorMessage.value = '';
  feedbackMessage.value = '';
}

function currentRouteKey() {
  return `${route.params.id}:${route.params.no}`;
}

function isActiveViewRequest(viewVersion, routeKey) {
  return viewVersion === activeViewVersion && routeKey === currentRouteKey();
}

function setChapterStatusOverride(status) {
  chapterStatusOverride.value = status;

  if (isTerminalChapterStatus(status)) {
    chapterTerminalStatusLock.value = status;
    return;
  }

  chapterTerminalStatusLock.value = '';
}

function shouldApplySnapshotChapterStatus(snapshotStatus) {
  if (!snapshotStatus) {
    return false;
  }

  if (!chapterTerminalStatusLock.value) {
    return true;
  }

  return snapshotStatus === chapterTerminalStatusLock.value;
}

function isTerminalChapterStatus(status) {
  return status === 'done' || status === 'failed';
}

function safeChapterNo() {
  return chapterNo.value > 0 ? chapterNo.value : Number(chapterData.value.chapter?.chapter_no || 0);
}

function createEmptyChapterData() {
  return {
    novel: null,
    state: 'invalid',
    chapter: {
      chapter_no: 0,
      title: '',
      summary: '',
      status: 'invalid',
      rewrite_count: 0,
    },
    content: '',
    placeholder: null,
    navigation: {
      prev: null,
      next: null,
    },
    comic: {
      page_count: 0,
      image_urls: [],
      pages: [],
    },
  };
}

function extractErrorMessage(error, fallback) {
  return error?.message || error?.error || fallback;
}

function formatChapterStatus(status, state) {
  if (status === 'writing') return '生成中';
  if (status === 'done') return '已完成';
  if (status === 'failed') return '生成失败';
  if (state === 'planned') return '待生成';
  if (state === 'invalid') return '不存在';
  return status || state || '-';
}

function formatComicStatus(status) {
  if (status === 'generating') return '生成中';
  if (status === 'done') return '已完成';
  if (status === 'failed') return '生成失败';
  if (status === 'invalid') return '不存在';
  return '待生成';
}

function formatComicFeedback(status) {
  if (status === 'generating') return `第 ${safeChapterNo()} 章漫画生成中`;
  if (status === 'done') return `第 ${safeChapterNo()} 章漫画已更新`;
  if (status === 'failed') return `第 ${safeChapterNo()} 章漫画生成失败`;
  return '漫画状态已更新';
}
</script>
