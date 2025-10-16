package service

import (
	"sync"
	"time"
)

// InMemoryIdempotencyService - временная реализация в памяти
// В production следует использовать Redis или базу данных
type InMemoryIdempotencyService struct {
	mu    sync.RWMutex
	store map[string]idempotencyRecord
}

type idempotencyRecord struct {
	userId     int
	resourceId int
	expiresAt  time.Time
}

func NewIdempotencyService() *InMemoryIdempotencyService {
	return &InMemoryIdempotencyService{
		store: make(map[string]idempotencyRecord),
	}
}

func (s *InMemoryIdempotencyService) CheckIdempotency(userId int, key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, exists := s.store[key]
	if !exists {
		return 0, nil
	}

	// Проверяем не устарела ли запись
	if time.Now().After(record.expiresAt) {
		delete(s.store, key)
		return 0, nil
	}

	// Проверяем принадлежит ли запись пользователю
	if record.userId != userId {
		return 0, nil
	}

	return record.resourceId, nil
}

func (s *InMemoryIdempotencyService) StoreIdempotency(userId int, key string, resourceId int, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[key] = idempotencyRecord{
		userId:     userId,
		resourceId: resourceId,
		expiresAt:  time.Now().Add(ttl),
	}

	// Очистка устаревших записей (в production это должно быть в отдельной горутине)
	for k, record := range s.store {
		if time.Now().After(record.expiresAt) {
			delete(s.store, k)
		}
	}

	return nil
}
