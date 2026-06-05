import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../stores/auth';

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/LoginView.vue'),
    meta: { guest: true },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/RegisterView.vue'),
    meta: { guest: true },
  },
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/HomeView.vue'),
    meta: { auth: true },
  },
  {
    path: '/novel/:id',
    name: 'NovelDetail',
    component: () => import('../views/NovelDetail.vue'),
    meta: { auth: true },
  },
  {
    path: '/novel/:id/chapter/:no',
    name: 'ChapterView',
    component: () => import('../views/ChapterView.vue'),
    meta: { auth: true },
  },
  {
    path: '/ai-configs',
    name: 'AIConfigs',
    component: () => import('../views/AIConfigView.vue'),
    meta: { auth: true },
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach(async (to) => {
  const auth = useAuthStore();
  if (!auth.user) await auth.fetchMe();

  if (to.meta.auth && !auth.user) return '/login';
  if (to.meta.guest && auth.user) return '/';
});

export default router;
