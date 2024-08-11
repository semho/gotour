package cond_semaphore

import (
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

type CondSemaphore struct {
	Capacity  int64
	Available int64
	Mu        *sync.Mutex
	Cond      *sync.Cond
}

func NewCondSemaphore(capacity int64) Semaphore {
	mu := &sync.Mutex{}
	s := &CondSemaphore{
		Capacity:  capacity,
		Available: capacity,
		Mu:        mu,
		Cond:      sync.NewCond(mu),
	}

	return s
}

func (s *CondSemaphore) Acquire(ctx context.Context, n int64) error {
	if n <= 0 {
		return fmt.Errorf("n must be positive")
	}

	s.Mu.Lock()
	defer s.Mu.Unlock()

	for s.Available < n {
		done := make(chan struct{})
		go func() {
			select {
			case <-ctx.Done():
				s.Cond.Broadcast() // разбудим ожидающих, чтобы они могли проверить контекст
			case <-done:
			}
		}()

		s.Cond.Wait()
		close(done)

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	s.Available -= n
	return nil
}

func (s *CondSemaphore) TryAcquire(n int64) bool {
	if n <= 0 {
		return false
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if s.Available >= n {
		s.Available -= n
		return true
	}
	return false
}

func (s *CondSemaphore) Release(n int64) {
	if n <= 0 {
		return
	}

	s.Mu.Lock()
	s.Available = min(s.Available+n, s.Capacity)
	s.Mu.Unlock()
	s.Cond.Broadcast()
}

func main() {
	start := time.Now()

	results := []int{10, 15, 8, 3, 17, 20, 1, 6, 10, 9, 13, 19}

	var wg sync.WaitGroup
	sem := NewCondSemaphore(6)
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
