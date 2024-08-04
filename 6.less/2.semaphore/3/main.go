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

type condSemaphore struct {
	capacity  int64
	available int64
	mu        *sync.Mutex
	cond      *sync.Cond
}

func NewCondSemaphore(capacity int64) Semaphore {
	mu := &sync.Mutex{}
	s := &condSemaphore{
		capacity:  capacity,
		available: capacity,
		mu:        mu,
		cond:      sync.NewCond(mu),
	}

	return s
}

func (s *condSemaphore) Acquire(ctx context.Context, n int64) error {
	if n <= 0 {
		return fmt.Errorf("n must be positive")
	}

	for {
		if s.TryAcquire(n) { //захват ресурсов, если есть
			return nil
		}

		s.mu.Lock()
		// если ресурсы недоступны, ждем
		waitCh := make(chan struct{})
		go func() {
			s.mu.Lock()
			s.cond.Wait()
			s.mu.Unlock()
			close(waitCh)
		}()

		s.mu.Unlock()
		select {
		case <-waitCh:
			select {
			case <-ctx.Done():
				s.mu.Lock()
				s.cond.Signal() // пробуждение горутины, ожидающей на cond.Wait()
				s.mu.Unlock()
				return ctx.Err()
			default:
				return nil //все ок, канал закрыт, будет захват
			}
		case <-ctx.Done():
			s.mu.Lock()
			s.cond.Signal()
			s.mu.Unlock()
			return ctx.Err()
		}
	}
}

func (s *condSemaphore) TryAcquire(n int64) bool {
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

func (s *condSemaphore) Release(n int64) {
	if n <= 0 {
		return
	}

	s.mu.Lock()
	s.available = min(s.available+n, s.capacity)
	s.mu.Unlock()
	s.cond.Broadcast()
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
