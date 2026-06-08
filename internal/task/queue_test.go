package task

import (
	"testing"
	"time"
)

func TestQueueCloseIsSafeToCallRepeatedly(t *testing.T) {
	processed := make(chan struct{}, 1)
	queue := New(func(Task) error {
		processed <- struct{}{}
		return nil
	}, func(Task) error {
		return nil
	})

	queue.EnqueueWrite(Task{NovelID: 1, ChapterNo: 1})

	select {
	case <-processed:
	case <-time.After(2 * time.Second):
		t.Fatal("expected queued task to be processed before shutdown")
	}

	queue.Close()
	queue.Close()
}
