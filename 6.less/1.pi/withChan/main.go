package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func rowLeibniz(start, step, intervals int, resultCh chan float64, doneChan chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	var partSum float64

	for i := start; i < intervals; i += step {
		//time.Sleep(1 * time.Second) //для теста syscall.SIGINT
		select {
		case <-doneChan:
			return
		default:
			if i == 0 {
				partSum = 1 //первый член ряда
				continue    // переходим к другой итерации, т.к. учтен
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
	doneChan := make(chan struct{})
	wg := sync.WaitGroup{}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for i := 0; i < num; i++ {
		wg.Add(1)
		go rowLeibniz(i, num, intervals, resultCh, doneChan, &wg)
	}

	go func() {
		<-sigChan
		fmt.Println("сигнал для закрытия горутины")
		close(doneChan)
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

	fmt.Println("done with chan")
	fmt.Printf("Схождение Pi: %.15f\n", pi)
}
