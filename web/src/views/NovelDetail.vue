<template>
  <div class="min-h-screen bg-gray-50">
    <nav class="bg-white shadow-sm border-b px-6 py-3 flex items-center justify-between">
      <div class="flex items-center gap-4">
        <router-link to="/" class="text-gray-400 hover:text-indigo-600">&larr; 返回</router-link>
        <h1 class="text-lg font-bold text-gray-800">{{ novel?.title || '加载中...' }}</h1>
      </div>
      <div class="flex items-center gap-3">
        <span class="text-xs text-gray-400">模式: {{ novel?.mode }} | 状态: {{ novel?.status }}</span>
        <span class="text-xs px-2 py-0.5 rounded-full" :class="sseConnected ? 'bg-emerald-100 text-emerald-700' : 'bg-gray-100 text-gray-500'">
          {{ sseConnected ? 'SSE已连接' : 'SSE未连接' }}
        </span>
      </div>
    </nav>

    <main class="max-w-4xl mx-auto px-4 py-6">
      <div v-if="loading" class="text-center text-gray-400 py-20">加载中...</div>
      <template v-else-if="novel">
        <div class="flex gap-3 mb-6">
          <button v-if="novel.status === 'drafting'"
            @click="handleStop" :disabled="acting"
            class="bg-red-100 text-red-600 px-4 py-1.5 rounded-lg text-sm hover:bg-red-200 disabled:opacity-50">
            停止生成
          </button>
          <button v-if="novel.status === 'stopped'"
            @click="handleResume" :disabled="acting"
            class="bg-indigo-100 text-indigo-600 px-4 py-1.5 rounded-lg text-sm hover:bg-indigo-200 disabled:opacity-50">
            继续生成
          </button>
        </div>

        <p v-if="progressMsg" class="text-sm text-indigo-600 mb-4">{{ progressMsg }}</p>

        <h2 class="font-semibold mb-3">章节列表</h2>
        <div v-if="chapters.length === 0" class="text-gray-400 text-sm">暂无章节</div>
        <div v-else class="space-y-2">
          <div v-for="ch in chapters" :key="ch.chapter_no"
            class="bg-white rounded-lg border p-4 flex items-center justify-between hover:shadow-sm transition-shadow">
            <div>
              <span class="font-medium">第 {{ ch.chapter_no }} 章</span>
              <span class="text-sm text-gray-500 ml-2">{{ ch.title }}</span>
            </div>
            <div class="flex gap-2 items-center">
              <span class="text-xs" :class="ch.status === 'done' ? 'text-green-500' : 'text-amber-500'">
                {{ ch.status === 'done' ? '已完成' : '生成中' }}
              </span>
              <button v-if="ch.status === 'done'"
                @click="handleRegen(ch.chapter_no)" :disabled="acting"
                class="text-xs text-indigo-500 hover:text-indigo-700 disabled:opacity-50">
                重生成
              </button>
            </div>
          </div>
        </div>

        <p v-if="errMsg" class="text-red-500 text-sm mt-4">{{ errMsg }}</p>
      </template>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue';
import { useRoute } from 'vue-router';
import { createNovelSSE, novelAPI } from '../api';

const route = useRoute();
const novel = ref(null);
const chapters = ref([]);
const loading = ref(true);
const acting = ref(false);
const errMsg = ref('');
const progressMsg = ref('');
const sseConnected = ref(false);
let eventSource = null;

async function loadDetail() {
  const detail = await novelAPI.detail(route.params.id);
  novel.value = detail.data.novel;
  chapters.value = detail.data.chapters || [];
}

function connectSSE() {
  eventSource = createNovelSSE(route.params.id);
  sseConnected.value = true;
  eventSource.addEventListener('progress', async (event) => {
    try {
      const payload = JSON.parse(event.data);
      if (payload.type === 'progress') {
        progressMsg.value = payload.detail || '';
      }
      if (payload.type === 'chapter' || payload.type === 'comic') {
        await loadDetail();
      }
      if (payload.type === 'error') {
        errMsg.value = payload.msg || '任务执行失败';
      }
    } catch {
      progressMsg.value = event.data;
    }
  });
  eventSource.onerror = () => {
    sseConnected.value = false;
    progressMsg.value = progressMsg.value || '实时连接已断开';
  };
}

onMounted(async () => {
  try {
    await loadDetail();
    connectSSE();
  } catch (e) {
    errMsg.value = e.message || '加载失败';
  } finally {
    loading.value = false;
  }
});

onBeforeUnmount(() => {
  if (eventSource) eventSource.close();
  sseConnected.value = false;
});

async function handleStop() {
  acting.value = true;
  try {
    await novelAPI.stop(route.params.id);
    await loadDetail();
  } catch (e) {
    errMsg.value = e.message;
  } finally {
    acting.value = false;
  }
}

async function handleResume() {
  acting.value = true;
  try {
    await novelAPI.resume(route.params.id);
    await loadDetail();
  } catch (e) {
    errMsg.value = e.message;
  } finally {
    acting.value = false;
  }
}

async function handleRegen(chapterNo) {
  acting.value = true;
  try {
    await novelAPI.regenChapter(route.params.id, chapterNo);
    progressMsg.value = `第 ${chapterNo} 章已重新入队`;
  } catch (e) {
    errMsg.value = e.message;
  } finally {
    acting.value = false;
  }
}
</script>
