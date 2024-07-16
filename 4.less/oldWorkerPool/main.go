package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Job struct {
	id   int
	time time.Duration
}

type WorkerPool struct {
	numWorkers int
	jobQueue   chan *Job
	maxError   int
	countError int
	quitCh     chan struct{}
	mu         *sync.Mutex
	wg         *sync.WaitGroup
}

func NewWorkPool(numWorkers, queueSize, maxError int) *WorkerPool {
	if maxError <= 0 {
		maxError = int(^uint(0) >> 1)
	}

	return &WorkerPool{
		numWorkers: numWorkers,
		jobQueue:   make(chan *Job, queueSize),
		maxError:   maxError,
		quitCh:     make(chan struct{}),
		mu:         &sync.Mutex{},
		wg:         &sync.WaitGroup{},
	}
}

func (wp *WorkerPool) Start() {
	for w := 1; w <= wp.numWorkers; w++ {
		wp.wg.Add(1)
		go wp.worker(w)
	}
}

func (wp *WorkerPool) SendJobs(jobs []time.Duration) {
	for i, dur := range jobs {
		wp.jobQueue <- &Job{id: i + 1, time: dur}
	}
	close(wp.jobQueue) //закрываем канал отправки
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	for job := range wp.jobQueue {
		select {
		case <-wp.quitCh: //канал закрыт, была ошибка, выходим
			return
		default:
			wp.processJob(id, job)
		}
	}
}

func (wp *WorkerPool) processJob(id int, job *Job) {
	wp.addError(job) //имитируем ошибку
	if wp.isLimitError() {
		wp.closeByLimitError()
		return
	}
	fmt.Printf("worker %d старт job %d время %d секунд\n", id, job.id, int64(job.time/time.Second))
	time.Sleep(job.time) //имитируем работу
	fmt.Printf("worker %d конец job %d время %d секунд\n", id, job.id, int64(job.time/time.Second))
}

func (wp *WorkerPool) addError(job *Job) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	if job.id == 3 || job.id == 4 {
		wp.countError++
	}
}

func (wp *WorkerPool) isLimitError() bool {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.countError >= wp.maxError
}

func (wp *WorkerPool) closeByLimitError() {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	if !wp.isQuitChClosed() { //чтобы не закрыть больше одного раза
		close(wp.quitCh)
	}
}

func (wp *WorkerPool) isQuitChClosed() bool {
	select {
	case <-wp.quitCh:
		return true
	default:
		return false
	}
}

func (wp *WorkerPool) Run(jobs []time.Duration) error {
	wp.Start()
	wp.SendJobs(jobs)
	wp.wg.Wait() //ждем все задачи
	if wp.isLimitError() {
		return ErrErrorsLimitExceeded
	}
	return nil
}

func main() {
	timeForJobs := []time.Duration{
		1 * time.Second,
		3 * time.Second,
		2 * time.Second,
		1 * time.Second,
		7 * time.Second,
		4 * time.Second,
		2 * time.Second,
	}
	wp := NewWorkPool(2, len(timeForJobs), 0)
	err := wp.Run(timeForJobs)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
