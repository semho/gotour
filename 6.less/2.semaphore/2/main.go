package main

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Semaphore interface {
	Acquire(context.Context, int64) error
	TryAcquire(int64) bool
	Release(int64)
}

type listSemaphore struct {
	capacity  int64
	available int64
	mu        *sync.Mutex
	waitList  *list.List
}

type waiter struct {
	n  int64
	ch chan struct{}
}

func NewListSemaphore(capacity int64) Semaphore {
	return &listSemaphore{
		capacity:  capacity,
		available: capacity,
		mu:        &sync.Mutex{},
		waitList:  list.New(),
	}
}

func (s *listSemaphore) Acquire(ctx context.Context, n int64) error {
	if n <= 0 {
		return fmt.Errorf("n must be positive")
	}

	if s.TryAcquire(n) {
		return nil
	}

	s.mu.Lock()
	w := &waiter{n: n, ch: make(chan struct{})}
	elem := s.waitList.PushBack(w)
	s.mu.Unlock()

	//сложная структура из-за невозможности дать приоритет отмены контекста
	select {
	case <-w.ch:
		select {
		case <-ctx.Done():
			s.mu.Lock()
			s.waitList.Remove(elem)
			s.mu.Unlock()
			return ctx.Err()
		default:
			return nil //все ок, канал закрыт, ресурс освобожден
		}
	case <-ctx.Done():
		s.mu.Lock()
		s.waitList.Remove(elem)
		s.mu.Unlock()
		return ctx.Err()
	}
}

func (s *listSemaphore) TryAcquire(n int64) bool {
	if n <= 0 {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.available >= n {
		s.available -= n
		return true
	}
	return false
}

func (s *listSemaphore) Release(n int64) {
	if n <= 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.waitList.Len() == 0 {
		s.available = s.capacity
		return
	}

	for n > 0 && s.waitList.Len() > 0 {
		e := s.waitList.Front()
		w := e.Value.(*waiter)

		s.waitList.Remove(e)
		s.available += w.n
		n--

		close(w.ch)

		if s.available > s.capacity {
			s.available = s.capacity
		}
	}

	if n > 0 {
		s.available = min(s.available+n, s.capacity)
	}
}

func main() {
	start := time.Now()

	results := []int{10, 15, 8, 3, 17, 20, 1, 6, 10, 9, 13, 19}

	var wg sync.WaitGroup
	sem := NewListSemaphore(6)
	var responses []int
	mu := &sync.Mutex{}

	for _, d := range results {
		wg.Add(1)
		go func(wait int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := sem.Acquire(ctx, 1); err != nil {
				log.Printf("Failed to acquire semaphore: %v", err)
				return
			}
			defer sem.Release(1)

			log.Printf("Waiting for %d seconds\n", wait)
			time.Sleep(time.Second * time.Duration(wait))
			log.Printf("Finished waiting for %d seconds\n", wait)

			mu.Lock()
			responses = append(responses, wait/2)
			mu.Unlock()
		}(d)
	}
	wg.Wait()

	for _, r := range responses {
		log.Printf("Got result %d", r)
	}

	log.Printf("Total time taken: %s\n", time.Since(start))
}
