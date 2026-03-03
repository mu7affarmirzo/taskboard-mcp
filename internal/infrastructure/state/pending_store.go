package state

import (
	"sync"
	"time"
)

type PendingTask struct {
	Title       string
	Description string
	DueDate     *time.Time
	Priority    string
	Labels      []string
	Checklist   []string
	Members     []string
	RawMessage  string
	CreatedAt   time.Time
}

type PendingStore struct {
	mu    sync.RWMutex
	tasks map[int64]PendingTask
}

func NewPendingStore() *PendingStore {
	return &PendingStore{
		tasks: make(map[int64]PendingTask),
	}
}

func (s *PendingStore) Set(telegramID int64, task PendingTask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task.CreatedAt = time.Now()
	s.tasks[telegramID] = task
}

func (s *PendingStore) Get(telegramID int64) (PendingTask, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[telegramID]
	return task, ok
}

func (s *PendingStore) Delete(telegramID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, telegramID)
}
