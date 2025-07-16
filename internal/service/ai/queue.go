package ai

import (
	"sync"
	"comic_video/internal/domain/entity"
)

type TaskQueue interface {
	Enqueue(task *entity.Task) error
	StartWorker(workerNum int, handler func(*entity.Task))
}

type MemoryTaskQueue struct {
	queue chan *entity.Task
	mu    sync.Mutex
}

func NewMemoryTaskQueue(size int) *MemoryTaskQueue {
	return &MemoryTaskQueue{queue: make(chan *entity.Task, size)}
}

func (q *MemoryTaskQueue) Enqueue(task *entity.Task) error {
	q.queue <- task
	return nil
}

func (q *MemoryTaskQueue) StartWorker(workerNum int, handler func(*entity.Task)) {
	for i := 0; i < workerNum; i++ {
		go func() {
			for task := range q.queue {
				handler(task)
			}
		}()
	}
} 