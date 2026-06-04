import { defineStore } from 'pinia';
import { ref } from 'vue';
import { authAPI } from '../api';

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null);
  const loading = ref(false);

  async function fetchMe() {
    try {
      const res = await authAPI.me();
      user.value = res.data;
    } catch {
      user.value = null;
    }
  }

  async function login(username, password) {
    loading.value = true;
    try {
      const res = await authAPI.login({ username, password });
      user.value = res.data;
      return res;
    } finally {
      loading.value = false;
    }
  }

  async function register(username, password, confirmPassword) {
    loading.value = true;
    try {
      const res = await authAPI.register({
        username,
        password,
        confirm_password: confirmPassword,
      });
      user.value = res.data;
      return res;
    } finally {
      loading.value = false;
    }
  }

  async function logout() {
    await authAPI.logout();
    user.value = null;
  }

  return { user, loading, fetchMe, login, register, logout };
});
