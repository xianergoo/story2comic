import axios from 'axios';

const DEFAULT_NOVEL_SSE_EVENT_NAMES = ['progress'];

const api = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
});

api.interceptors.response.use(
  (res) => {
    const data = res.data;
    if (data.code !== 0 && data.code !== undefined) {
      return Promise.reject(data);
    }
    return data;
  },
  (err) => {
    const msg = err?.response?.data?.message || err.message || '网络错误';
    return Promise.reject({ code: -1, error: msg, message: msg });
  }
);

export default api;

export const authAPI = {
  login: (data) => api.post('/auth/login', data),
  register: (data) => api.post('/auth/register', data),
  logout: () => api.post('/auth/logout'),
  me: () => api.get('/auth/me'),
};

export function parseNovelSSEMessage(rawEvent) {
  const rawData = rawEvent?.data;

  if (typeof rawData !== 'string' || rawData.trim() === '') {
    throw new Error('Invalid SSE payload');
  }

  const parsed = JSON.parse(rawData);
  if (!parsed || typeof parsed !== 'object' || typeof parsed.type !== 'string') {
    throw new Error('Invalid SSE payload shape');
  }

  const { type, payload, ...rest } = parsed;

  let normalizedPayload = null;
  if (Object.prototype.hasOwnProperty.call(parsed, 'payload')) {
    normalizedPayload = payload;
  } else if (Object.keys(rest).length > 0) {
    normalizedPayload = rest;
  }

  return {
    type,
    payload: normalizedPayload,
    raw: parsed,
  };
}

export function createNovelSSE(novelId) {
  return new EventSource(`/api/v1/sse?novel_id=${novelId}`, { withCredentials: true });
}

function normalizeNovelSSEEventNames(eventNames) {
  if (eventNames === undefined) {
    return DEFAULT_NOVEL_SSE_EVENT_NAMES;
  }

  if (!Array.isArray(eventNames)) {
    return [];
  }

  return [...new Set(eventNames.filter((eventName) => typeof eventName === 'string' && eventName))];
}

export function bindNovelSSEMessages(eventSource, listener, options = {}) {
  const eventNames = normalizeNovelSSEEventNames(options.eventNames);

  const wrappedListener = (event) => {
    let message;

    try {
      message = parseNovelSSEMessage(event);
    } catch (error) {
      options.onParseError?.(error, event);
      return;
    }

    listener(message, event);
  };

  eventSource.onmessage = wrappedListener;
  eventNames.forEach((eventName) => {
    eventSource.addEventListener(eventName, wrappedListener);
  });

  return () => {
    if (eventSource.onmessage === wrappedListener) {
      eventSource.onmessage = null;
    }
    eventNames.forEach((eventName) => {
      eventSource.removeEventListener(eventName, wrappedListener);
    });
  };
}

export const novelAPI = {
  list: () => api.get('/novels'),
  create: (data) => api.post('/novels', data),
  detail: (id) => api.get(`/novels/${id}`),
  stop: (id) => api.post(`/novels/${id}/stop`),
  resume: (id) => api.post(`/novels/${id}/resume`),
  regenChapter: (novelId, chapterNo) =>
    api.post(`/novels/${novelId}/chapters/${chapterNo}/regenerate`),
};

export const chapterAPI = {
  view: (novelId, chapterNo) =>
    api.get(`/novels/${novelId}/chapters/${chapterNo}`),
};

export const agentAPI = {
  list: (novelId) => api.get('/agent/tasks', { params: { novel_id: novelId } }),
  create: (data) => api.post('/agent/tasks', data),
  detail: (id) => api.get(`/agent/tasks/${id}`),
  cancel: (id) => api.post(`/agent/tasks/${id}/cancel`),
  checkpoints: (id) => api.get(`/agent/tasks/${id}/checkpoints`),
  updateCheckpoint: (id, data) => api.put(`/agent/checkpoints/${id}`, data),
};

export const aiConfigAPI = {
  list: () => api.get('/ai-configs'),
  create: (data) => api.post('/ai-configs', data),
  update: (id, data) => api.put(`/ai-configs/${id}`, data),
  remove: (id) => api.delete(`/ai-configs/${id}`),
};
