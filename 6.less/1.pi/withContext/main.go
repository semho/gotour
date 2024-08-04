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

func rowLeibniz(ctx context.Context, start, step, intervals int, resultCh chan float64, wg *sync.WaitGroup) {
	defer wg.Done()
	var partSum float64

	for i := start; i < intervals; i += step {
		//time.Sleep(1 * time.Second) //для теста syscall.SIGINT
		select {
		case <-ctx.Done():
			return
		default:
			if i == 0 {
				partSum = 1 //первый член ряда
				continue    //переходим к другой итерации, т.к. учтен
			}
			den := float64(i)*2 + 1
			frac := 1 / den
			if i%2 != 0 {
				frac = -frac
			}
			partSum += frac
		}
	}
	resultCh <- partSum
}

func main() {
	var num int
	var intervals int
	flag.IntVar(&num, "n", 1, "количество горутин")
	flag.IntVar(&intervals, "i", 100, "количество повторений цикла")
	flag.Parse()

	resultCh := make(chan float64, num)
	wg := sync.WaitGroup{}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for i := 0; i < num; i++ {
		wg.Add(1)
		go rowLeibniz(ctx, i, num, intervals, resultCh, &wg)
	}

	go func() {
		<-sigChan
		fmt.Println("сигнал для закрытия горутины")
		cancel()
		wg.Wait()
		close(resultCh)
		os.Exit(0)
	}()

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var pi float64
	for partSum := range resultCh {
		pi += partSum
	}
	pi *= 4

	fmt.Println("done with context")
	fmt.Printf("Схождение Pi: %.15f\n", pi)
}
