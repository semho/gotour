package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func rowLeibniz(ctx context.Context, start, step int, resultCh chan float64, wg *sync.WaitGroup) {
	defer wg.Done()
	var partSum float64
	for i := start; ; i += step {
		select {
		case <-ctx.Done():
			resultCh <- partSum
			return
		default:
			frac := 4.0 / float64(2*i+1)
			if i%2 != 0 {
				partSum -= frac
			} else {
				partSum += frac
			}
		}
	}
}

func main() {
	var num int
	flag.IntVar(&num, "n", 1, "количество горутин")
	flag.Parse()

	resultCh := make(chan float64, num)
	wg := sync.WaitGroup{}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < num; i++ {
		wg.Add(1)
		go rowLeibniz(ctx, i, num, resultCh, &wg)
	}

	go func() {
		<-sigChan
		fmt.Println("сигнал для закрытия горутины")
		cancel()
	}()

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	pi := 0.0
	for partSum := range resultCh {
		pi += partSum
	}

	fmt.Println("done with context")
	fmt.Printf("Схождение Pi: %.15f\n", pi)
}
