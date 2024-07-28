package main

import (
	"fmt"
	"sync"
	"time"
)

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	out := in
	for _, stage := range stages {
		out = workPipeline(out, done, stage)
	}
	return out
}

func workPipeline(ch In, done In, stage Stage) Out {
	workCh := stage(ch)
	out := make(Bi)

	go func() {
		defer close(out)
		for val := range workCh {
			select {
			case <-done:
				return
			default:
				select {
				case out <- val:
				}
			}
		}
	}()

	return out
}

func WorkStage(in In) (out Out) {
	ch := make(Bi)
	out = ch

	go func() {
		defer close(ch)
		for val := range in {
			time.Sleep(100 * time.Millisecond)
			ch <- val.(int) * 2
		}
	}()

	return out
}

func main() {
	start := time.Now()
	in := make(Bi)
	done := make(Bi)
	countCh := 10
	go func() {
		defer close(in)
		for i := 1; i <= countCh; i++ {
			in <- i
		}
	}()

	result := ExecutePipeline(in, done, WorkStage, WorkStage, WorkStage, WorkStage)
	var results []int
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		time.Sleep(600 * time.Millisecond)
		close(done)
	}()

	go func() {
		defer wg.Done()
		for val := range result {
			results = append(results, val.(int))
		}
	}()

	wg.Wait()

	if len(results) == countCh {
		for _, val := range results {
			fmt.Println(val)
		}
	} else {
		fmt.Println("Пайплайн не выполнен полностью.")
	}

	duration := time.Since(start)
	fmt.Printf("Время выполнения пайплайна: %v\n", duration)
}
