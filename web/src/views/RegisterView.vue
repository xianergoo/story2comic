<template>
  <div class="min-h-screen bg-gray-50 flex items-center justify-center">
    <div class="w-full max-w-sm bg-white rounded-xl shadow p-8">
      <h1 class="text-2xl font-bold text-center text-indigo-600 mb-6">注册</h1>
      <form @submit.prevent="handleRegister" class="space-y-4">
        <input v-model="form.username" placeholder="用户名" required
          class="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-400 outline-none" />
        <input v-model="form.password" type="password" placeholder="密码" required
          class="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-400 outline-none" />
        <input v-model="form.confirmPassword" type="password" placeholder="确认密码" required
          class="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-400 outline-none" />
        <p v-if="form.confirmPassword && form.password !== form.confirmPassword"
          class="text-red-500 text-sm">两次密码不一致</p>
        <p v-if="error" class="text-red-500 text-sm">{{ error }}</p>
        <button :disabled="auth.loading"
          class="w-full bg-indigo-600 text-white py-2 rounded-lg hover:bg-indigo-700 font-medium disabled:opacity-50">
          {{ auth.loading ? '注册中...' : '注册' }}
        </button>
      </form>
      <p class="text-center text-sm text-gray-500 mt-4">
        已有账号？
        <router-link to="/login" class="text-indigo-600">登录</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../stores/auth';

const auth = useAuthStore();
const router = useRouter();
const form = ref({ username: '', password: '', confirmPassword: '' });
const error = ref('');

async function handleRegister() {
  error.value = '';
  if (form.value.password !== form.value.confirmPassword) {
    error.value = '两次密码不一致';
    return;
  }
  try {
    await auth.register(form.value.username, form.value.password, form.value.confirmPassword);
    router.push('/');
  } catch (e) {
    error.value = e.message || '注册失败';
  }
}
</script>
