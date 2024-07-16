package workerpool

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

type WorkerPool struct {
	wg             *sync.WaitGroup
	mu             *sync.Mutex
	closeOnce      *sync.Once
	tasksChan      chan Task
	doneChan       chan struct{}
	errorCount     int
	maxCountErrors int
	tasksCount     int
	workerCount    int
}

func NewWorkerPool(n, m int) *WorkerPool {
	return &WorkerPool{
		wg:             &sync.WaitGroup{},
		mu:             &sync.Mutex{},
		closeOnce:      &sync.Once{},
		tasksChan:      make(chan Task),
		doneChan:       make(chan struct{}),
		maxCountErrors: m,
		workerCount:    n,
	}
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for {
		select {
		case task, ok := <-wp.tasksChan:
			if !ok {
				return
			}
			err := task()
			wp.mu.Lock()
			if err != nil {
				wp.errorCount++
				if wp.errorCount >= wp.maxCountErrors {
					wp.Stop()
					wp.mu.Unlock()
					return
				}
			}
			wp.tasksCount++
			wp.mu.Unlock()
		case <-wp.doneChan:
			return
		}
	}
}

func (wp *WorkerPool) Start(tasks []Task) error {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}

	go func() {
		defer close(wp.tasksChan)
		for _, task := range tasks {
			select {
			case wp.tasksChan <- task:
			case <-wp.doneChan:
				return
			}
		}
	}()

	wp.wg.Wait()
	wp.Stop()

	wp.mu.Lock()
	defer wp.mu.Unlock()
	if wp.errorCount >= wp.maxCountErrors {
		return ErrErrorsLimitExceeded
	}

	return nil
}

func (wp *WorkerPool) Stop() {
	wp.closeOnce.Do(
		func() {
			close(wp.doneChan)
		},
	)
}

func Run(tasks []Task, n, m int) error {
	if m <= 0 {
		return ErrErrorsLimitExceeded
	}

	wp := NewWorkerPool(n, m)
	return wp.Start(tasks)
}
