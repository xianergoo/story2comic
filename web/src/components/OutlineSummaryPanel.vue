<template>
  <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
    <div class="flex items-center justify-between gap-3">
      <div>
        <h2 class="text-lg font-semibold text-slate-900">大纲摘要</h2>
        <p class="mt-1 text-sm text-slate-500">接入 `outline_stream`，展示空态、生成中与完成态。</p>
      </div>
      <span class="rounded-full px-3 py-1 text-xs font-medium" :class="stateClass">
        {{ stateLabel }}
      </span>
    </div>

    <div v-if="state === 'empty'" class="mt-4 rounded-2xl border border-dashed border-slate-200 px-4 py-8 text-center text-sm text-slate-500">
      还没有可展示的大纲内容
    </div>

    <div v-else class="mt-4 space-y-4">
      <div class="rounded-2xl bg-slate-50 p-4">
        <pre class="whitespace-pre-wrap break-words text-sm leading-6 text-slate-700">{{ content }}</pre>
      </div>

      <div v-if="outline?.chapter_plan?.length" class="rounded-2xl border border-slate-200 p-4">
        <p class="text-sm font-medium text-slate-700">章节规划</p>
        <ul class="mt-3 space-y-2 text-sm text-slate-600">
          <li v-for="item in outline.chapter_plan" :key="item.chapter_no">
            第 {{ item.chapter_no }} 章 · {{ item.title || '未命名章节' }}
          </li>
        </ul>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  outline: {
    type: Object,
    default: null,
  },
  streamContent: {
    type: String,
    default: '',
  },
  streamDone: {
    type: Boolean,
    default: false,
  },
});

const content = computed(() => props.streamContent || props.outline?.content || '');

const state = computed(() => {
  if (!content.value) {
    return 'empty';
  }
  return props.streamDone || props.outline?.content ? 'done' : 'streaming';
});

const stateLabel = computed(() => {
  if (state.value === 'empty') return '空态';
  if (state.value === 'streaming') return '生成中';
  return '已完成';
});

const stateClass = computed(() => {
  if (state.value === 'empty') return 'bg-slate-100 text-slate-600';
  if (state.value === 'streaming') return 'bg-amber-100 text-amber-700';
  return 'bg-emerald-100 text-emerald-700';
});
</script>
