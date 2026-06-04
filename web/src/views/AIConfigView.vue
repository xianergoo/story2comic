<template>
  <div class="min-h-screen bg-gray-50">
    <nav class="bg-white shadow-sm border-b px-6 py-3 flex items-center justify-between">
      <div class="flex items-center gap-4">
        <router-link to="/" class="text-gray-400 hover:text-indigo-600">&larr; 返回</router-link>
        <h1 class="text-lg font-bold text-gray-800">AI 配置</h1>
      </div>
    </nav>

    <main class="max-w-5xl mx-auto px-4 py-6">
      <section class="bg-white rounded-xl border p-5 mb-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="font-semibold">{{ editingId ? '编辑配置' : '新增配置' }}</h2>
          <button v-if="editingId" @click="resetForm" class="text-sm text-gray-500 hover:text-gray-700">取消编辑</button>
        </div>
        <form @submit.prevent="handleSubmit" class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <input v-model="form.name" placeholder="配置名称" required class="px-3 py-2 border rounded-lg" />
          <input v-model="form.provider" placeholder="provider，如 custom/openai/qwen" required class="px-3 py-2 border rounded-lg" />
          <input v-model="form.api_key" placeholder="API Key" required class="px-3 py-2 border rounded-lg" />
          <input v-model="form.base_url" placeholder="Base URL" class="px-3 py-2 border rounded-lg" />
          <input v-model="form.text_model" placeholder="文本模型" required class="px-3 py-2 border rounded-lg" />
          <input v-model="form.image_model" placeholder="图像模型" class="px-3 py-2 border rounded-lg" />
          <label class="flex items-center gap-2 text-sm text-gray-600">
            <input v-model="form.is_default" type="checkbox" /> 设为默认
          </label>
          <div class="md:col-span-2">
            <button :disabled="saving" class="bg-indigo-600 text-white px-4 py-2 rounded-lg disabled:opacity-50">
              {{ saving ? '保存中...' : editingId ? '更新配置' : '保存配置' }}
            </button>
          </div>
        </form>
        <p v-if="error" class="text-red-500 text-sm mt-3">{{ error }}</p>
      </section>

      <section>
        <h2 class="font-semibold mb-3">已有配置</h2>
        <div v-if="loading" class="text-sm text-gray-400">加载中...</div>
        <div v-else-if="configs.length === 0" class="text-sm text-gray-400">暂无配置</div>
        <div v-else class="space-y-3">
          <div v-for="cfg in configs" :key="cfg.id" class="bg-white rounded-xl border p-4 flex items-center justify-between gap-4">
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2 mb-1">
                <span class="font-medium">{{ cfg.name }}</span>
                <span v-if="cfg.is_default" class="text-xs bg-indigo-100 text-indigo-700 px-2 py-0.5 rounded-full">默认</span>
              </div>
              <p class="text-sm text-gray-500 truncate">{{ cfg.provider }} / {{ cfg.text_model }}</p>
              <p class="text-xs text-gray-400 truncate">{{ cfg.base_url }}</p>
            </div>
            <div class="flex items-center gap-3 shrink-0">
              <button v-if="!cfg.is_default" @click="handleSetDefault(cfg)" class="text-sm text-indigo-500 hover:text-indigo-600">设默认</button>
              <button @click="handleEdit(cfg)" class="text-sm text-gray-500 hover:text-gray-700">编辑</button>
              <button @click="handleDelete(cfg.id)" class="text-sm text-red-500 hover:text-red-600">删除</button>
            </div>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue';
import { aiConfigAPI } from '../api';

const configs = ref([]);
const loading = ref(true);
const saving = ref(false);
const error = ref('');
const editingId = ref(null);
const form = ref({
  name: '',
  provider: 'custom',
  api_key: '',
  base_url: '',
  text_model: '',
  image_model: '',
  is_default: false,
});

async function loadConfigs() {
  loading.value = true;
  try {
    const res = await aiConfigAPI.list();
    configs.value = res.data || [];
  } finally {
    loading.value = false;
  }
}

function resetForm() {
  editingId.value = null;
  form.value = {
    name: '',
    provider: 'custom',
    api_key: '',
    base_url: '',
    text_model: '',
    image_model: '',
    is_default: false,
  };
}

function handleEdit(cfg) {
  editingId.value = cfg.id;
  form.value = {
    name: cfg.name,
    provider: cfg.provider,
    api_key: cfg.api_key,
    base_url: cfg.base_url,
    text_model: cfg.text_model,
    image_model: cfg.image_model,
    is_default: cfg.is_default,
  };
}

async function handleSubmit() {
  error.value = '';
  saving.value = true;
  try {
    if (editingId.value) {
      await aiConfigAPI.update(editingId.value, form.value);
    } else {
      await aiConfigAPI.create(form.value);
    }
    resetForm();
    await loadConfigs();
  } catch (e) {
    error.value = e.message || '保存失败';
  } finally {
    saving.value = false;
  }
}

async function handleSetDefault(cfg) {
  await aiConfigAPI.update(cfg.id, {
    name: cfg.name,
    provider: cfg.provider,
    api_key: cfg.api_key,
    base_url: cfg.base_url,
    text_model: cfg.text_model,
    image_model: cfg.image_model,
    is_default: true,
  });
  await loadConfigs();
}

async function handleDelete(id) {
  await aiConfigAPI.remove(id);
  await loadConfigs();
}

onMounted(loadConfigs);
</script>
