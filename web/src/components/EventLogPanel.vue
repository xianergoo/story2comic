<template>
  <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
    <div class="flex items-center justify-between gap-3">
      <div>
        <h2 class="text-lg font-semibold text-slate-900">事件流</h2>
        <p class="mt-1 text-sm text-slate-500">仅展示 `log` 与 `error` 事件时间线。</p>
      </div>
      <span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-600">
        {{ items.length }} 条
      </span>
    </div>

    <div v-if="items.length === 0" class="mt-4 rounded-2xl border border-dashed border-slate-200 px-4 py-8 text-center text-sm text-slate-500">
      还没有事件流记录
    </div>

    <ol v-else class="mt-4 space-y-3">
      <li
        v-for="item in items"
        :key="item.id"
        class="rounded-2xl border px-4 py-3"
        :class="item.level === 'error' ? 'border-rose-200 bg-rose-50' : 'border-slate-200 bg-slate-50'"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="item.level === 'error' ? 'bg-rose-100 text-rose-700' : 'bg-slate-200 text-slate-700'">
                {{ item.level === 'error' ? '错误' : '日志' }}
              </span>
              <span v-if="item.step" class="text-xs font-medium text-slate-500">{{ item.step }}</span>
              <span v-if="item.chapter_no" class="text-xs text-slate-500">第 {{ item.chapter_no }} 章</span>
            </div>
            <p class="mt-2 text-sm text-slate-700">{{ item.message }}</p>
          </div>
          <time class="shrink-0 text-xs text-slate-400">{{ item.createdAt }}</time>
        </div>
      </li>
    </ol>
  </section>
</template>

<script setup>
defineProps({
  items: {
    type: Array,
    default: () => [],
  },
});
</script>
