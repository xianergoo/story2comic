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
        <div v-else class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          <router-link v-for="n in novels" :key="n.id" :to="`/novel/${n.id}`"
            class="bg-white rounded-lg shadow-sm border hover:shadow-md transition-shadow p-4 block">
            <div class="flex items-center justify-between mb-1">
              <span class="text-xs px-2 py-0.5 rounded-full"
                :class="n.mode === 'inspiration' ? 'bg-indigo-100 text-indigo-700' : n.mode === 'outline' ? 'bg-amber-100 text-amber-700' : 'bg-emerald-100 text-emerald-700'">
                {{ n.mode === 'inspiration' ? '灵感' : n.mode === 'outline' ? '章纲' : '盲盒' }}
              </span>
              <span class="text-xs" :class="statusClass(n.status)">{{ n.status }}</span>
            </div>
            <h3 class="font-semibold mb-1 truncate">{{ n.title }}</h3>
            <p class="text-xs text-gray-400">{{ n.chapter_count ? n.chapter_count + ' 章' : '待生成' }}</p>
          </router-link>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useAuthStore } from '../stores/auth';
import { useRouter } from 'vue-router';
import { novelAPI } from '../api';

const auth = useAuthStore();
const router = useRouter();
const novels = ref([]);
const loading = ref(true);
const creating = ref(false);
const createError = ref('');
const createForm = ref({
  title: '',
  summary: '',
  mode: 'inspiration',
  image_mode: 'single',
  ai_config_id: null,
});

onMounted(async () => {
  try {
    const res = await novelAPI.list();
    novels.value = res.data || [];
  } catch (e) {
    console.error(e);
  } finally {
    loading.value = false;
  }
});

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

function statusClass(s) {
  return {
    drafting: 'text-indigo-500',
    completed: 'text-green-500',
    stopped: 'text-gray-400',
    failed: 'text-red-500',
  }[s] || 'text-indigo-500';
}
</script>
