package main

import (
	"fmt"
	"testing"
	"time"
)

func TestWorkerPoolConcurrency(t *testing.T) {
	t.Run(
		"Normal operation", func(t *testing.T) {
			numJobs := 10
			numWorkers := 3
			maxErrors := 5

			wp := NewWorkPool(numWorkers, numJobs, maxErrors)

			jobs := make([]time.Duration, numJobs)
			for i := range jobs {
				jobs[i] = time.Nanosecond
			}

			err := wp.Run(jobs)

			// не получили ошибку
			if err != nil {
				t.Errorf("Expected nil error, got %v", err)
			} else {
				fmt.Println("Success: all jobs processed without exceeding error limit")
			}

			// все задачи были выполнены
			if len(wp.jobQueue) != 0 {
				t.Error("Expected all jobs to be processed")
			}

			// количество ошибок
			expectedErrors := 2 // Задачи 3 и 4 вызывают ошибки
			if wp.countError != expectedErrors {
				t.Errorf("Expected %d errors, got %d", expectedErrors, wp.countError)
			}
		},
	)
	t.Run(
		"Error limit exceeded", func(t *testing.T) {
			numJobs := 5
			numWorkers := 2
			maxErrors := 1

			wp := NewWorkPool(numWorkers, numJobs, maxErrors)

			jobs := make([]time.Duration, numJobs)
			for i := range jobs {
				jobs[i] = time.Nanosecond
			}

			err := wp.Run(jobs)

			// есть ошибка
			if err == ErrErrorsLimitExceeded {
				fmt.Printf("Success: received expected error: %v\n", err)
			} else {
				t.Errorf("Expected ErrErrorsLimitExceeded, got %v", err)
			}

			if wp.countError != maxErrors {
				t.Errorf("Expected %d errors, got %d", maxErrors, wp.countError)
			}
		},
	)
}
