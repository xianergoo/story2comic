<template>
  <section class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h2 class="text-lg font-semibold text-slate-900">漫画分镜</h2>
        <p class="mt-1 text-sm text-slate-500">展示本章已生成图片，并响应漫画生成状态变化。</p>
      </div>
      <span class="rounded-full px-3 py-1 text-xs font-medium" :class="badgeClass">
        {{ statusLabel }}
      </span>
    </div>

    <div v-if="message" class="mt-4 rounded-2xl border border-indigo-200 bg-indigo-50 px-4 py-3 text-sm text-indigo-700">
      {{ message }}
    </div>

    <div v-if="pages.length > 0" class="mt-4 space-y-4">
      <article v-for="page in pages" :key="page.page_no" class="rounded-2xl border border-slate-200 bg-slate-50 p-4">
        <div class="flex items-center justify-between gap-3">
          <h3 class="text-sm font-semibold text-slate-800">第 {{ page.page_no }} 页</h3>
          <span class="text-xs text-slate-500">{{ page.panel_count || page.image_urls.length || 0 }} 格</span>
        </div>

        <div class="mt-3 grid gap-3 sm:grid-cols-2">
          <div
            v-for="(imageUrl, index) in page.image_urls"
            :key="`${page.page_no}-${index}-${imageUrl}`"
            class="overflow-hidden rounded-2xl bg-slate-200"
          >
            <img :src="imageUrl" :alt="`第 ${page.page_no} 页图片 ${index + 1}`" class="h-full w-full object-cover" loading="lazy">
          </div>
        </div>
      </article>
    </div>

    <div v-else-if="imageUrls.length > 0" class="mt-4 grid gap-3 sm:grid-cols-2">
      <div
        v-for="(imageUrl, index) in imageUrls"
        :key="`${index}-${imageUrl}`"
        class="overflow-hidden rounded-2xl bg-slate-200"
      >
        <img :src="imageUrl" :alt="`漫画图片 ${index + 1}`" class="h-full w-full object-cover" loading="lazy">
      </div>
    </div>

    <div v-else class="mt-4 rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-5 py-10 text-center">
      <p class="text-base font-medium text-slate-700">{{ emptyTitle }}</p>
      <p class="mt-2 text-sm leading-6 text-slate-500">{{ emptyMessage }}</p>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  pages: {
    type: Array,
    default: () => [],
  },
  imageUrls: {
    type: Array,
    default: () => [],
  },
  status: {
    type: String,
    default: 'planned',
  },
  state: {
    type: String,
    default: 'invalid',
  },
  message: {
    type: String,
    default: '',
  },
});

const statusLabel = computed(() => {
  if (props.status === 'generating') return '生成中';
  if (props.status === 'done') return '已完成';
  if (props.status === 'failed') return '生成失败';
  if (props.state === 'invalid') return '不存在';
  return '待生成';
});

const badgeClass = computed(() => {
  if (props.status === 'generating') return 'bg-amber-100 text-amber-700';
  if (props.status === 'done') return 'bg-emerald-100 text-emerald-700';
  if (props.status === 'failed') return 'bg-rose-100 text-rose-700';
  if (props.state === 'invalid') return 'bg-rose-100 text-rose-700';
  return 'bg-slate-200 text-slate-700';
});

const emptyTitle = computed(() => {
  if (props.state === 'invalid') return '没有可展示的漫画章节';
  if (props.status === 'generating') return '漫画生成中';
  if (props.status === 'failed') return '漫画生成失败';
  return '漫画尚未生成';
});

const emptyMessage = computed(() => {
  if (props.state === 'invalid') {
    return '当前章节不可用，因此漫画区域也没有可渲染内容。';
  }
  if (props.status === 'generating') {
    return '本章漫画正在生成，新的图片完成后会自动更新到这里。';
  }
  if (props.status === 'failed') {
    return '漫画流程未成功完成，你可以重生成本章后再次查看。';
  }
  return '正文完成后，会在这里展示对应的分镜图片。';
});
</script>
