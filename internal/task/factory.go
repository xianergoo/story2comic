package task

import (
	"fmt"
)

// TaskFactory 任务工厂，创建任务处理器
type TaskFactory struct {
	handler TaskHandler
	publisher SSEPublisher
	config   ConfigProvider
}

// NewTaskFactory 创建任务工厂
func NewTaskFactory(
	handler TaskHandler,
	publisher SSEPublisher,
	config ConfigProvider,
) *TaskFactory {
	return &TaskFactory{
		handler:   handler,
		publisher: publisher,
		config:    config,
	}
}

// CreateQueue 创建任务队列
func (f *TaskFactory) CreateQueue() *Queue {
	queue := New(f.createWriteHandler(), f.createImageHandler())
	
	// 设置图片任务入队回调
	queue.SetOnEnqueueImage(func(t Task) {
		if f.handler != nil {
			f.handler.OnEnqueueImage(t)
		}
	})
	
	return queue
}

func (f *TaskFactory) createWriteHandler() func(Task) error {
	return func(t Task) error {
		if f.handler == nil {
			return fmt.Errorf("任务处理器未设置")
		}
		return f.handler.HandleWrite(t)
	}
}

func (f *TaskFactory) createImageHandler() func(Task) error {
	return func(t Task) error {
		if f.handler == nil {
			return fmt.Errorf("任务处理器未设置")
		}
		return f.handler.HandleImage(t)
	}
}