package render

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"comic_video/internal/domain/entity"
)

// TaskQueue 通用任务队列接口
// 支持多类型任务的入队、消费、worker启动等
// 可用于渲染、AI、视频等任务统一调度

type TaskQueue interface {
	Enqueue(task *entity.Task) error
	StartWorker(workerNum int, handler func(*entity.Task))
}

// MemoryTaskQueue 内存实现
// 支持多类型任务

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

// --- 保留原有渲染队列实现，便于平滑迁移 ---
type RenderQueue interface {
	Enqueue(renderID uuid.UUID) error
	StartWorker(workerNum int, service Service)
}

type MemoryRenderQueue struct {
	queue chan uuid.UUID
}

func NewMemoryRenderQueue(size int) *MemoryRenderQueue {
	return &MemoryRenderQueue{queue: make(chan uuid.UUID, size)}
}

func (q *MemoryRenderQueue) Enqueue(id uuid.UUID) error {
	q.queue <- id
	return nil
}

func (q *MemoryRenderQueue) StartWorker(workerNum int, service Service) {
	for i := 0; i < workerNum; i++ {
		go func() {
			for id := range q.queue {
				_ = service.ProcessRender(context.Background(), id)
			}
		}()
	}
} 