package handler

import (
	"io"
	"sync"
	"github.com/gin-gonic/gin"
)

type SSEHandler struct {
	mu      sync.RWMutex
	clients map[uint]chan string
}

func NewSSEHandler() *SSEHandler {
	return &SSEHandler{clients: make(map[uint]chan string)}
}

func (h *SSEHandler) Subscribe(c *gin.Context) {
	novelID := c.GetUint("novel_id")
	ch := make(chan string, 10)

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
		msg, ok := <-ch
		if !ok {
			return false
		}
		c.SSEvent("progress", msg)
		return true
	})
}

func (h *SSEHandler) Push(novelID uint, message string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if ch, ok := h.clients[novelID]; ok {
		select {
		case ch <- message:
		default:
		}
	}
}
