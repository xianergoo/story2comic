<template>
  <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
    <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
      <div>
        <p class="text-sm font-medium text-slate-500">进度总览</p>
        <div class="mt-3 grid grid-cols-2 gap-3 sm:grid-cols-4">
          <article
            v-for="item in stats"
            :key="item.key"
            class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3"
          >
            <p class="text-xs uppercase tracking-wide text-slate-500">{{ item.label }}</p>
            <p class="mt-2 text-2xl font-semibold text-slate-900">{{ item.value }}</p>
          </article>
        </div>
      </div>

      <div class="min-w-0 rounded-2xl border border-indigo-100 bg-indigo-50 px-4 py-4 lg:w-80">
        <p class="text-sm font-medium text-indigo-700">当前章节</p>
        <div v-if="currentChapter" class="mt-2 space-y-3">
          <div>
            <p class="text-lg font-semibold text-slate-900">
              第 {{ currentChapter.chapter_no }} 章
              <span v-if="currentChapter.title" class="ml-1 text-base font-medium text-slate-600">
                {{ currentChapter.title }}
              </span>
            </p>
            <p class="mt-1 text-sm text-slate-600">
              状态: {{ formatStatus(currentChapter.status, currentChapter.state) }}
            </p>
          </div>

          <button
            v-if="cta"
            type="button"
            class="inline-flex items-center rounded-full bg-slate-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:bg-slate-300"
            :disabled="cta.disabled"
            @click="$emit('cta')"
          >
            {{ cta.label }}
          </button>
        </div>

        <p v-else class="mt-2 text-sm text-slate-500">暂无当前章节，等待大纲或章节计划。</p>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  progress: {
    type: Object,
    default: () => ({}),
  },
  currentChapter: {
    type: Object,
    default: null,
  },
  cta: {
    type: Object,
    default: null,
  },
});

defineEmits(['cta']);

const stats = computed(() => [
  { key: 'planned', label: 'planned', value: props.progress?.planned_count ?? 0 },
  { key: 'generated', label: 'generated', value: props.progress?.generated_count ?? 0 },
  { key: 'text_done', label: 'text_done', value: props.progress?.text_done_count ?? 0 },
  { key: 'remaining', label: 'remaining', value: props.progress?.remaining_count ?? 0 },
]);

function formatStatus(status, state) {
  if (status === 'done') return '已完成';
  if (status === 'writing') return '生成中';
  if (status === 'failed') return '失败';
  if (status === 'planned') return state === 'generated' ? '待完成' : '待生成';
  if (status) return status;
  if (state === 'planned') return '待生成';
  return '未知';
}
</script>
