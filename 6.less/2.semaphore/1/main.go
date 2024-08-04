package main

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

type bufferSemaphore struct {
	capacity int64
	tokens   chan struct{}
	mu       *sync.Mutex
}

func NewSemaphore(capacity int64) Semaphore {
	return &bufferSemaphore{
		capacity: capacity,
		tokens:   make(chan struct{}, capacity),
		mu:       &sync.Mutex{},
	}
}

func (s *bufferSemaphore) Acquire(ctx context.Context, n int64) error {
	if n <= 0 {
		return fmt.Errorf("n must be positive")
	}

	for {
		if s.TryAcquire(n) {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}

func (s *bufferSemaphore) TryAcquire(n int64) bool {
	if n <= 0 {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if int64(len(s.tokens))+n > s.capacity {
		return false
	}

	for i := int64(0); i < n; i++ {
		select {
		case s.tokens <- struct{}{}:
		default:
			for ; i > 0; i-- {
				<-s.tokens
			}
			return false
		}
	}
	return true
}

func (s *bufferSemaphore) Release(n int64) {
	if n <= 0 {
		return
	}
	for i := int64(0); i < n; i++ {
		<-s.tokens
	}
}

func main() {
	start := time.Now()

	results := []int{10, 15, 8, 3, 17, 20, 1, 6, 10, 9, 13, 19}

	var wg sync.WaitGroup
	sem := NewSemaphore(6)
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