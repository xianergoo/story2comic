package task

import "fmt"

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
	writeChan chan Task
	imageChan chan Task
	writeFunc func(Task) error
	imageFunc func(Task) error
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

func (q *Queue) EnqueueWrite(t Task) { q.writeChan <- t }
func (q *Queue) EnqueueImage(t Task) { q.imageChan <- t }

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
