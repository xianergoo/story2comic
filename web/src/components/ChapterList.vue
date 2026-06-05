<template>
  <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
    <div class="flex items-center justify-between gap-3">
      <div>
        <h2 class="text-lg font-semibold text-slate-900">章节列表</h2>
        <p class="mt-1 text-sm text-slate-500">已接入快照状态、当前高亮与重生成入口。</p>
      </div>
      <span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-600">
        {{ chapters.length }} 章
      </span>
    </div>

    <div v-if="chapters.length === 0" class="mt-4 rounded-2xl border border-dashed border-slate-200 px-4 py-8 text-center text-sm text-slate-500">
      暂无章节快照
    </div>

    <div v-else class="mt-4 space-y-3">
      <article
        v-for="chapter in chapters"
        :key="chapter.chapter_no"
        class="rounded-2xl border px-4 py-4 transition"
        :class="chapter.chapter_no === currentChapterNo ? 'border-indigo-300 bg-indigo-50 shadow-sm' : 'border-slate-200 bg-slate-50/60'"
      >
        <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <h3 class="text-base font-semibold text-slate-900">第 {{ chapter.chapter_no }} 章</h3>
              <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="badgeClass(chapter)">
                {{ formatStatus(chapter.status, chapter.state) }}
              </span>
              <span
                v-if="chapter.chapter_no === currentChapterNo"
                class="rounded-full bg-indigo-600 px-2.5 py-1 text-xs font-medium text-white"
              >
                当前生成中
              </span>
            </div>

            <p v-if="chapter.title" class="mt-2 text-sm font-medium text-slate-700">{{ chapter.title }}</p>
            <p v-if="chapter.summary" class="mt-1 line-clamp-2 text-sm text-slate-500">{{ chapter.summary }}</p>

            <div class="mt-3 flex flex-wrap gap-3 text-xs text-slate-500">
              <span>正文: {{ chapter.has_content ? '已写入' : '未写入' }}</span>
              <span>重写次数: {{ chapter.rewrite_count ?? 0 }}</span>
              <span>
                漫画: {{ formatComicStatus(chapter.comic?.status) }}
                <template v-if="chapter.comic?.page_count">
                  ({{ chapter.comic?.done_page_count ?? 0 }}/{{ chapter.comic?.page_count }})
                </template>
              </span>
            </div>
          </div>

          <div class="flex shrink-0 flex-wrap items-center gap-2">
            <a
              v-if="canRead(chapter)"
              class="inline-flex items-center rounded-full border border-slate-300 bg-white px-3 py-1.5 text-sm font-medium text-slate-700 transition hover:border-slate-400 hover:text-slate-900"
              :href="readHref(chapter.chapter_no)"
            >
              阅读
            </a>
            <button
              type="button"
              class="inline-flex items-center rounded-full border border-indigo-200 bg-indigo-50 px-3 py-1.5 text-sm font-medium text-indigo-700 transition hover:bg-indigo-100 disabled:cursor-not-allowed disabled:opacity-50"
              :disabled="regenerating || chapter.chapter_no === currentChapterNo && isBusy"
              @click="$emit('regenerate', chapter.chapter_no)"
            >
              重生成
            </button>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup>
const props = defineProps({
  novelId: {
    type: [String, Number],
    required: true,
  },
  chapters: {
    type: Array,
    default: () => [],
  },
  currentChapterNo: {
    type: Number,
    default: 0,
  },
  regenerating: {
    type: Boolean,
    default: false,
  },
  isBusy: {
    type: Boolean,
    default: false,
  },
});

defineEmits(['regenerate']);

function canRead(chapter) {
  return chapter?.state === 'generated' || chapter?.has_content;
}

function readHref(chapterNo) {
  return `/novel/${props.novelId}/chapter/${chapterNo}`;
}

function badgeClass(chapter) {
  if (chapter.status === 'done') return 'bg-emerald-100 text-emerald-700';
  if (chapter.status === 'writing') return 'bg-amber-100 text-amber-700';
  if (chapter.status === 'failed') return 'bg-rose-100 text-rose-700';
  if (chapter.state === 'planned') return 'bg-slate-200 text-slate-700';
  return 'bg-sky-100 text-sky-700';
}

function formatStatus(status, state) {
  if (status === 'done') return '已完成';
  if (status === 'writing') return '生成中';
  if (status === 'failed') return '失败';
  if (status === 'planned') return state === 'generated' ? '待完成' : '待生成';
  if (status) return status;
  if (state === 'planned') return '待生成';
  return '未知';
}

function formatComicStatus(status) {
  if (status === 'done') return '已完成';
  if (status === 'pending') return '生成中';
  if (status === 'planned') return '未开始';
  return status || '未知';
}
</script>
