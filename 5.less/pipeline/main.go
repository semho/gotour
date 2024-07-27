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
			case out <- val:
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

	go func() {
		defer close(in)
		for i := 1; i <= 5; i++ {
			in <- i
		}
	}()

	result := ExecutePipeline(in, done, WorkStage, WorkStage, WorkStage, WorkStage)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		time.Sleep(600 * time.Millisecond)
		close(done)
	}()

	go func() {
		defer wg.Done()
		for val := range result {
			fmt.Println(val)
		}
	}()

	wg.Wait()

	duration := time.Since(start)
	fmt.Printf("Время выполнения пайплайна: %v\n", duration)
}
