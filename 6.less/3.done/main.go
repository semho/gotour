package main

import (
	"fmt"
	"sync"
	"time"
)

func or(channels ...<-chan interface{}) <-chan interface{} {
	count := len(channels)
	output := make(chan interface{})

	var once sync.Once
	var wg sync.WaitGroup
	wg.Add(count)

	for _, channel := range channels {
		go func(ch <-chan interface{}) {
			defer wg.Done()
			for {
				select {
				case _, outputOk := <-output:
					if !outputOk {
						return
					}
				default:
					select {
					case val, ok := <-ch:
						if !ok {
							closeOnce(&once, output)
							return
						}
						output <- val
					}
				}
			}
		}(channel)
	}

	go func() {
		wg.Wait()
		closeOnce(&once, output)
	}()

	return output
}

func closeOnce(once *sync.Once, output chan interface{}) {
	once.Do(
		func() {
			close(output)
		},
	)
}

func main() {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)

	fmt.Printf("done after %v", time.Since(start)) // ~1 second
}
