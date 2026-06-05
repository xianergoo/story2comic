import { computed, onBeforeUnmount, ref } from 'vue';
import { bindNovelSSEMessages, createNovelSSE } from '../api';

export function useNovelStream(options = {}) {
  const {
    novelId,
    handlers = {},
    onEvent,
    onError,
    onReconnectSnapshot,
    eventNames,
  } = options;

  const status = ref('idle');
  const lastError = ref(null);

  let eventSource = null;
  let unbindMessages = null;
  let hasConnected = false;
  let pendingReconnectSnapshot = false;

  const isConnected = computed(() => status.value === 'open');

  function cleanupListeners() {
    if (unbindMessages) {
      unbindMessages();
      unbindMessages = null;
    }

    if (eventSource) {
      eventSource.onopen = null;
      eventSource.onerror = null;
    }
  }

  function close() {
    cleanupListeners();

    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }

    status.value = 'closed';
  }

  function dispatchEvent(message, rawEvent) {
    onEvent?.(message, rawEvent);

    const handler = handlers[message.type];
    if (handler) {
      handler(message.payload, message, rawEvent);
    }
  }

  function handleStreamError(error) {
    lastError.value = error;
    onError?.(error);
  }

  function open(overrideNovelId) {
    const resolvedNovelId = overrideNovelId ?? novelId;
    if (!resolvedNovelId) {
      const error = new Error('novelId is required to open stream');
      handleStreamError(error);
      throw error;
    }

    close();

    status.value = 'connecting';
    lastError.value = null;
    eventSource = createNovelSSE(resolvedNovelId);

    unbindMessages = bindNovelSSEMessages(
      eventSource,
      (message, rawEvent) => {
        try {
          dispatchEvent(message, rawEvent);
        } catch (error) {
          handleStreamError(error);
        }
      },
      {
        eventNames,
        onParseError: (error) => {
          handleStreamError(error);
        },
      }
    );

    eventSource.onopen = async () => {
      status.value = 'open';

      if (hasConnected && pendingReconnectSnapshot && onReconnectSnapshot) {
        pendingReconnectSnapshot = false;
        try {
          await onReconnectSnapshot();
        } catch (error) {
          handleStreamError(error);
        }
      }

      hasConnected = true;
    };

    eventSource.onerror = () => {
      pendingReconnectSnapshot = hasConnected;
      status.value = eventSource?.readyState === EventSource.CLOSED ? 'closed' : 'error';
      handleStreamError(new Error('Novel SSE connection error'));
    };

    return eventSource;
  }

  onBeforeUnmount(() => {
    close();
  });

  return {
    status,
    isConnected,
    lastError,
    open,
    close,
  };
}
