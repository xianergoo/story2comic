package task

import (
	"fmt"
	"sync"
)

type TaskType string

const (
	TaskWrite TaskType = "write"
	TaskImage TaskType = "image"
)

type Task struct {
	NovelID    uint
	ChapterNo  int
	Type       TaskType
	Suggestion string
}

type Queue struct {
	writeChan      chan Task
	imageChan      chan Task
	writeFunc      func(Task) error
	imageFunc      func(Task) error
	onEnqueueImage func(Task) // 图片任务入队回调
	closeOnce      sync.Once
}

func New(writeFn, imageFn func(Task) error) *Queue {
	q := &Queue{
		writeChan: make(chan Task, 10),
		imageChan: make(chan Task, 10),
		writeFunc: writeFn,
		imageFunc: imageFn,
	}
	go q.writeWorker()
	go q.imageWorker()
	return q
}

// SetOnEnqueueImage 设置图片任务入队回调
func (q *Queue) SetOnEnqueueImage(callback func(Task)) {
	q.onEnqueueImage = callback
}

func (q *Queue) EnqueueWrite(t Task) { 
	t.Type = TaskWrite
	q.writeChan <- t 
}

func (q *Queue) EnqueueImage(t Task) { 
	t.Type = TaskImage
	q.imageChan <- t 
	if q.onEnqueueImage != nil {
		q.onEnqueueImage(t)
	}
}

func (q *Queue) Close() {
	q.closeOnce.Do(func() {
		close(q.writeChan)
		close(q.imageChan)
	})
}

func (q *Queue) writeWorker() {
	for t := range q.writeChan {
		fmt.Printf("TASK-WRITE: novel=%d ch=%d\n", t.NovelID, t.ChapterNo)
		if err := q.writeFunc(t); err != nil {
			fmt.Printf("write task failed: novel=%d ch=%d err=%v\n", t.NovelID, t.ChapterNo, err)
		}
	}
}

func (q *Queue) imageWorker() {
	for t := range q.imageChan {
		fmt.Printf("TASK-IMAGE: novel=%d ch=%d\n", t.NovelID, t.ChapterNo)
		if err := q.imageFunc(t); err != nil {
			fmt.Printf("image task failed: novel=%d ch=%d err=%v\n", t.NovelID, t.ChapterNo, err)
		}
	}
}
