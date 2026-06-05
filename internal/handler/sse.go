package handler

import (
	"encoding/json"
	"io"
	"novelforge/internal/task"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

type SSEHandler struct {
	mu      sync.RWMutex
	clients map[uint]chan task.Event
}

func NewSSEHandler() *SSEHandler {
	return &SSEHandler{clients: make(map[uint]chan task.Event)}
}

func (h *SSEHandler) Subscribe(c *gin.Context) {
	id, _ := strconv.Atoi(c.Query("novel_id"))
	novelID := uint(id)
	ch := make(chan task.Event, 10)

	h.mu.Lock()
	h.clients[novelID] = ch
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, novelID)
		h.mu.Unlock()
		close(ch)
	}()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	c.Stream(func(w io.Writer) bool {
		event, ok := <-ch
		if !ok {
			return false
		}

		payload, err := json.Marshal(event)
		if err != nil {
			payload = []byte(`{"type":"error","payload":{"message":"marshal_sse_event_failed"}}`)
		}

		c.SSEvent(task.SSETransportEventName, string(payload))
		return true
	})
}

func (h *SSEHandler) Push(novelID uint, message string) {
	h.PushEvent(novelID, task.EventFromLegacyMessage(message))
}

func (h *SSEHandler) PushEvent(novelID uint, event task.Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if ch, ok := h.clients[novelID]; ok {
		select {
		case ch <- event:
		default:
		}
	}
}
