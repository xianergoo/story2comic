<template>
  <div class="min-h-screen bg-gray-50">
    <nav class="bg-white shadow-sm border-b px-6 py-3 flex items-center justify-between">
      <h1 class="text-xl font-bold text-indigo-600">NovelForge</h1>
      <div class="flex items-center gap-4">
        <router-link to="/ai-configs" class="text-sm text-indigo-600 hover:text-indigo-700">AI配置</router-link>
        <span class="text-sm text-gray-500">{{ auth.user?.username }}</span>
        <button @click="handleLogout" class="text-sm text-gray-400 hover:text-red-500">退出</button>
      </div>
    </nav>

    <main class="max-w-6xl mx-auto px-4 py-6">
      <div class="mb-6">
        <h2 class="text-lg font-semibold mb-3">创建新小说</h2>
        <form @submit.prevent="handleCreate" class="flex gap-3 flex-wrap items-end">
          <div>
            <label class="text-xs text-gray-500 block mb-1">标题</label>
            <input v-model="createForm.title" placeholder="小说标题" required
              class="px-3 py-2 border rounded-lg text-sm w-48 focus:ring-2 focus:ring-indigo-400 outline-none" />
          </div>
          <div>
            <label class="text-xs text-gray-500 block mb-1">模式</label>
            <select v-model="createForm.mode"
              class="px-3 py-2 border rounded-lg text-sm w-32 focus:ring-2 focus:ring-indigo-400 outline-none">
              <option value="inspiration">灵感</option>
              <option value="outline">章纲</option>
              <option value="blindbox">盲盒</option>
            </select>
          </div>
          <div>
            <label class="text-xs text-gray-500 block mb-1">梗概</label>
            <input v-model="createForm.summary" placeholder="一句话梗概..."
              class="px-3 py-2 border rounded-lg text-sm w-64 focus:ring-2 focus:ring-indigo-400 outline-none" />
          </div>
          <button type="submit" :disabled="creating"
            class="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-indigo-700 disabled:opacity-50">
            {{ creating ? '创建中...' : '创建' }}
          </button>
        </form>
        <p v-if="createError" class="text-red-500 text-sm mt-2">{{ createError }}</p>
      </div>

      <div>
        <h2 class="text-lg font-semibold mb-3">我的小说</h2>
        <div v-if="loading" class="text-gray-400 text-sm">加载中...</div>
        <div v-else-if="novels.length === 0" class="text-gray-400 text-sm">还没有小说，创建一个吧</div>
        <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
          <article
            v-for="n in novels"
            :key="n.id"
            class="flex h-full flex-col rounded-2xl border border-gray-200 bg-white p-4 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md"
          >
            <div class="mb-3 flex items-start justify-between gap-3">
              <div class="min-w-0 flex-1">
                <h3 class="truncate text-base font-semibold text-gray-900">{{ n.title }}</h3>
                <p class="mt-1 text-xs text-gray-400">{{ chapterSummary(n) }}</p>
              </div>
              <span
                class="shrink-0 rounded-full px-2.5 py-1 text-xs font-medium"
                :class="modeClass(n.mode)"
              >
                {{ formatMode(n.mode) }}
              </span>
            </div>

            <dl class="grid grid-cols-2 gap-3 text-sm">
              <div class="rounded-xl bg-gray-50 px-3 py-2">
                <dt class="text-xs text-gray-400">总状态</dt>
                <dd class="mt-1 font-medium" :class="statusClass(n.status)">{{ formatNovelStatus(n.status) }}</dd>
              </div>
              <div class="rounded-xl bg-gray-50 px-3 py-2">
                <dt class="text-xs text-gray-400">写作状态</dt>
                <dd class="mt-1 font-medium" :class="subStatusClass(n.text_status)">{{ formatTextStatus(n.text_status) }}</dd>
              </div>
              <div class="rounded-xl bg-gray-50 px-3 py-2">
                <dt class="text-xs text-gray-400">图片状态</dt>
                <dd class="mt-1 font-medium" :class="subStatusClass(n.image_status)">{{ formatImageStatus(n.image_status) }}</dd>
              </div>
              <div class="rounded-xl bg-gray-50 px-3 py-2">
                <dt class="text-xs text-gray-400">已完成章节</dt>
                <dd class="mt-1 font-medium text-gray-700">{{ chapterSummary(n) }}</dd>
              </div>
            </dl>

            <p
              v-if="actionFeedback[n.id]"
              class="mt-3 rounded-xl border px-3 py-2 text-xs"
              :class="actionFeedback[n.id].type === 'error'
                ? 'border-red-200 bg-red-50 text-red-600'
                : 'border-indigo-200 bg-indigo-50 text-indigo-700'"
            >
              {{ actionFeedback[n.id].message }}
            </p>

            <p
              v-if="refreshFeedback[n.id]"
              class="mt-3 rounded-xl border px-3 py-2 text-xs"
              :class="refreshFeedback[n.id].type === 'error'
                ? 'border-amber-200 bg-amber-50 text-amber-700'
                : 'border-emerald-200 bg-emerald-50 text-emerald-700'"
            >
              {{ refreshFeedback[n.id].message }}
            </p>

            <div class="mt-4 flex flex-wrap gap-2">
              <button
                type="button"
                class="rounded-lg bg-slate-900 px-3 py-2 text-sm font-medium text-white transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:bg-slate-300"
                :disabled="resumeDisabled(n)"
                @click="handleResume(n)"
              >
                {{ actionState[n.id] === 'resume' ? '继续中...' : '快捷继续' }}
              </button>
              <button
                type="button"
                class="rounded-lg bg-rose-600 px-3 py-2 text-sm font-medium text-white transition hover:bg-rose-500 disabled:cursor-not-allowed disabled:bg-rose-200"
                :disabled="stopDisabled(n)"
                @click="handleStop(n)"
              >
                {{ actionState[n.id] === 'stop' ? '停止中...' : '快捷停止' }}
              </button>
              <button
                type="button"
                class="rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium text-gray-600 transition hover:border-indigo-200 hover:text-indigo-600"
                @click="goToDetail(n.id)"
              >
                查看详情
              </button>
            </div>
          </article>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue';
import { useAuthStore } from '../stores/auth';
import { useRouter } from 'vue-router';
import api, { novelAPI } from '../api';

const auth = useAuthStore();
const router = useRouter();
const novels = ref([]);
const loading = ref(true);
const creating = ref(false);
const createError = ref('');
const actionState = reactive({});
const actionFeedback = reactive({});
const refreshFeedback = reactive({});
const createForm = ref({
  title: '',
  summary: '',
  mode: 'inspiration',
  image_mode: 'single',
  ai_config_id: null,
});

onMounted(async () => {
  await loadNovels();
});

async function loadNovels() {
  try {
    const res = await novelAPI.list();
    novels.value = res.data || [];
  } catch (e) {
    console.error(e);
  } finally {
    loading.value = false;
  }
}

async function handleCreate() {
  createError.value = '';
  creating.value = true;
  try {
    const res = await novelAPI.create(createForm.value);
    createForm.value = {
      title: '',
      summary: '',
      mode: 'inspiration',
      image_mode: 'single',
      ai_config_id: null,
    };
    router.push(`/novel/${res.data.id}`);
  } catch (e) {
    createError.value = e.message || '创建失败';
  } finally {
    creating.value = false;
  }
}

async function handleLogout() {
  await auth.logout();
  router.push('/login');
}

function goToDetail(id) {
  router.push(`/novel/${id}`);
}

function setActionFeedback(id, type, message) {
  actionFeedback[id] = { type, message };
}

function setRefreshFeedback(id, type, message) {
  refreshFeedback[id] = { type, message };
}

function clearRefreshFeedback(id) {
  delete refreshFeedback[id];
}

function modeClass(mode) {
  return mode === 'inspiration'
    ? 'bg-indigo-100 text-indigo-700'
    : mode === 'outline'
      ? 'bg-amber-100 text-amber-700'
      : 'bg-emerald-100 text-emerald-700';
}

function formatMode(mode) {
  return mode === 'inspiration' ? '灵感' : mode === 'outline' ? '章纲' : '盲盒';
}

function formatNovelStatus(status) {
  if (status === 'drafting') return '生成中';
  if (status === 'completed') return '已完成';
  if (status === 'stopped') return '已停止';
  if (status === 'failed') return '失败';
  return status || '-';
}

function formatTextStatus(status) {
  if (status === 'writing') return '写作中';
  if (status === 'paused') return '已暂停';
  if (status === 'done') return '已完成';
  if (status === 'stopped') return '已停止';
  if (status === 'failed') return '失败';
  return status || '-';
}

function formatImageStatus(status) {
  if (status === 'generating' || status === 'drawing') return '生成中';
  if (status === 'paused') return '已暂停';
  if (status === 'done') return '已完成';
  if (status === 'stopped') return '已停止';
  if (status === 'failed') return '失败';
  return status || '未开始';
}

function chapterSummary(novel) {
  return novel.chapter_count ? `${novel.chapter_count} 章` : '0 章';
}

function statusClass(s) {
  return {
    drafting: 'text-indigo-500',
    completed: 'text-green-500',
    stopped: 'text-gray-400',
    failed: 'text-red-500',
  }[s] || 'text-indigo-500';
}

function subStatusClass(status) {
  return {
    writing: 'text-indigo-500',
    generating: 'text-indigo-500',
    drawing: 'text-indigo-500',
    paused: 'text-amber-600',
    done: 'text-green-500',
    stopped: 'text-gray-400',
    failed: 'text-red-500',
  }[status] || 'text-gray-700';
}

function isTextRunning(novel) {
  return novel.text_status === 'writing';
}

function isImageRunning(novel) {
  return novel.image_status === 'generating' || novel.image_status === 'drawing';
}

function isTextResumable(novel) {
  return novel.text_status === 'paused'
    || novel.text_status === 'stopped'
    || novel.text_status === 'failed';
}

function resumeDisabled(novel) {
  return Boolean(actionState[novel.id]) || novel.status === 'completed' || isTextRunning(novel) || !isTextResumable(novel);
}

function stopDisabled(novel) {
  return Boolean(actionState[novel.id]) || novel.status === 'completed' || !isTextRunning(novel);
}

function syncNovelSnapshot(updatedNovel) {
  const index = novels.value.findIndex((item) => item.id === updatedNovel.id);
  if (index === -1) {
    return;
  }
  novels.value[index] = {
    ...novels.value[index],
    ...updatedNovel,
  };
}

async function refreshNovel(id) {
  const detail = await novelAPI.detail(id);
  if (detail?.data?.novel) {
    syncNovelSnapshot(detail.data.novel);
  } else {
    await loadNovels();
  }
}

async function handleResume(novel) {
  actionState[novel.id] = 'resume';
  setActionFeedback(novel.id, 'info', '正在请求继续文本生成，成功后将进入详情页...');
  clearRefreshFeedback(novel.id);
  try {
    await api.post(`/novels/${novel.id}/resume`, { pipeline: 'text' });
    setActionFeedback(novel.id, 'info', '继续文本生成请求已发送。');
    try {
      await refreshNovel(novel.id);
      setRefreshFeedback(novel.id, 'success', '首页卡片已刷新。');
    } catch (refreshError) {
      setRefreshFeedback(novel.id, 'error', '首页卡片刷新失败，将进入详情页查看最新状态。');
      console.error(refreshError);
    }
    router.push(`/novel/${novel.id}`);
  } catch (e) {
    setActionFeedback(novel.id, 'error', e.message || e.error || '继续生成失败');
  } finally {
    actionState[novel.id] = '';
  }
}

async function handleStop(novel) {
  actionState[novel.id] = 'stop';
  setActionFeedback(novel.id, 'info', '正在请求停止文本生成...');
  clearRefreshFeedback(novel.id);
  try {
    await api.post(`/novels/${novel.id}/stop`, { pipeline: 'text' });
    setActionFeedback(novel.id, 'info', '停止文本生成请求已发送。');
    try {
      await refreshNovel(novel.id);
      setRefreshFeedback(novel.id, 'success', '首页卡片已刷新。');
    } catch (refreshError) {
      setRefreshFeedback(novel.id, 'error', '首页卡片刷新失败，请进入详情页确认最新状态。');
      console.error(refreshError);
    }
  } catch (e) {
    setActionFeedback(novel.id, 'error', e.message || e.error || '停止生成失败');
  } finally {
    actionState[novel.id] = '';
  }
}
</script>
