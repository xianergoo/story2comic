<template>
  <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h2 class="text-lg font-semibold text-slate-900">章节正文</h2>
        <p class="mt-1 text-sm text-slate-500">支持已完成、待生成、生成中与失败态展示。</p>
      </div>
      <span class="rounded-full px-3 py-1 text-xs font-medium" :class="badgeClass">
        {{ statusLabel }}
      </span>
    </div>

    <div v-if="message" class="mt-4 rounded-2xl border border-indigo-200 bg-indigo-50 px-4 py-3 text-sm text-indigo-700">
      {{ message }}
    </div>

    <div v-if="error" class="mt-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
      {{ error }}
    </div>

    <div v-if="showPlaceholder" class="mt-4 rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-5 py-10 text-center">
      <p class="text-base font-medium text-slate-700">{{ placeholderTitle }}</p>
      <p class="mt-2 text-sm leading-6 text-slate-500">{{ placeholderMessage }}</p>
    </div>

    <div v-else class="mt-4 rounded-2xl bg-slate-950 px-5 py-5 text-slate-100">
      <pre class="whitespace-pre-wrap break-words font-sans text-[15px] leading-7 text-slate-100">{{ displayContent }}</pre>
      <div v-if="isStreaming" class="mt-4 flex items-center gap-2 text-sm text-amber-300">
        <span class="inline-flex h-2.5 w-2.5 rounded-full bg-amber-300" />
        正在持续接收正文片段
      </div>
      <div v-else-if="interrupted && hasContent" class="mt-4 text-sm text-slate-300">
        连接暂时中断，已保留当前内容，恢复后会自动回拉最新快照。
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  state: {
    type: String,
    default: 'invalid',
  },
  status: {
    type: String,
    default: '',
  },
  content: {
    type: String,
    default: '',
  },
  placeholder: {
    type: Object,
    default: null,
  },
  isStreaming: {
    type: Boolean,
    default: false,
  },
  interrupted: {
    type: Boolean,
    default: false,
  },
  message: {
    type: String,
    default: '',
  },
  error: {
    type: String,
    default: '',
  },
});

const hasContent = computed(() => props.content.trim().length > 0);
const showPlaceholder = computed(() => !hasContent.value && !props.isStreaming);
const displayContent = computed(() => {
  if (hasContent.value) {
    return props.content;
  }
  return '正文内容将在这里逐段展开...';
});

const statusLabel = computed(() => {
  if (props.isStreaming || props.status === 'writing') return '生成中';
  if (props.status === 'done') return '已完成';
  if (props.status === 'failed') return '生成失败';
  if (props.state === 'planned') return '待生成';
  if (props.state === 'invalid') return '不存在';
  return props.status || props.state || '未知';
});

const badgeClass = computed(() => {
  if (props.isStreaming || props.status === 'writing') return 'bg-amber-100 text-amber-700';
  if (props.status === 'done') return 'bg-emerald-100 text-emerald-700';
  if (props.status === 'failed') return 'bg-rose-100 text-rose-700';
  if (props.state === 'planned') return 'bg-slate-200 text-slate-700';
  if (props.state === 'invalid') return 'bg-rose-100 text-rose-700';
  return 'bg-sky-100 text-sky-700';
});

const placeholderTitle = computed(() => {
  if (props.state === 'invalid') return '该章节不存在';
  if (props.status === 'failed') return '本章生成失败';
  if (props.state === 'planned') return '本章尚未生成';
  return '正文暂不可用';
});

const placeholderMessage = computed(() => {
  if (props.placeholder?.message) {
    return props.placeholder.message;
  }
  if (props.status === 'failed') {
    return '你可以尝试点击“重生成本章”，重新触发正文与漫画流程。';
  }
  if (props.state === 'planned') {
    return '当前章节已进入规划，但正文还未落库。生成开始后，正文会在这里流式出现。';
  }
  return '请返回作品详情页检查章节规划，或切换到相邻章节继续阅读。';
});
</script>
