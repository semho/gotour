package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "string"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, strings.Split(value, ",")...) //теперь название файлов можно передать через запятую
	return nil
}

var (
	otherFile *os.File
	mu        sync.Mutex

	processedLines      map[string]int64
	errorLines          int64
	openInputFiles      int64
	totalProcessedLines int64
)

func init() {
	var err error
	otherFile, err = os.Create("other")
	if err != nil {
		log.Fatal("Ошибка открытия файла 'other':", err)
	}
	processedLines = make(map[string]int64)
}

func main() {
	defer closeFile(otherFile)

	var inputFiles arrayFlags
	var outputFile string
	flag.Var(&inputFiles, "inputs", "файлы для чтения через запятую")
	flag.StringVar(&outputFile, "output", "output", "название выходного файла")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	metricsCtx, metricsCancel := context.WithCancel(ctx)
	go reportMetrics(metricsCtx)

	if err := processFiles(ctx, inputFiles, outputFile); err != nil {
		if err != context.Canceled {
			log.Printf("Ошибка обработки файлов: %v", err)
		}
	}

	metricsCancel()

	//пауза, чтобы дать время завершиться reportMetrics, можно через wg сделать потом
	time.Sleep(100 * time.Millisecond)
	printFinalCounters()
}

func reportMetrics(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second) //как часто смотрим метрики
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			select {
			case <-ticker.C:
				printMetrics()
			}

		}
	}
}

func printMetrics() {
	fmt.Printf(
		"Метрики: Обработано строк: %d, Строк с ошибками: %d, Открыто входных файлов: %d\n",
		atomic.LoadInt64(&totalProcessedLines), atomic.LoadInt64(&errorLines), atomic.LoadInt64(&openInputFiles),
	)
}

func printFinalCounters() {
	fmt.Println("Финальные счетчики:")
	mu.Lock()
	for file, count := range processedLines {
		fmt.Printf("Файл %s: обработано строк %d\n", file, count)
	}
	mu.Unlock()
	fmt.Printf("Всего обработано строк: %d\n", atomic.LoadInt64(&totalProcessedLines))
	fmt.Printf("Всего строк с ошибками: %d\n", atomic.LoadInt64(&errorLines))
	fmt.Printf("Открытых входных файлов: %d\n", atomic.LoadInt64(&openInputFiles))
}

func processFiles(ctx context.Context, inputFiles []string, outputFile string) error {
	channels, err := processingFiles(ctx, inputFiles)
	if err != nil {
		return fmt.Errorf("ошибка при обработке входных файлов: %w", err)
	}

	mergedChannel := merge(ctx, channels...)

	return saveResult(ctx, outputFile, mergedChannel)
}

func saveResult(ctx context.Context, outputFile string, mergedChannel <-chan int) error {
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("ошибка создания выходного файла: %w", err)
	}

	defer closeFile(outFile)

	writer := bufio.NewWriter(outFile)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Получен сигнал завершения. Сохранение промежуточных результатов...")
			return nil
		default:
			select {
			case num, ok := <-mergedChannel:
				if !ok {
					return writer.Flush()
				}
				_, err = writer.WriteString(fmt.Sprintf("%d\n", num))
				if err != nil {
					return fmt.Errorf("ошибка записи в файл: %w", err)
				}
				err = writer.Flush()
				if err != nil {
					return fmt.Errorf("ошибка сбрасывания данных из буфера в выходной файл: %w", err)
				}
			}
		}
	}
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		fmt.Printf("Ошибка при закрытии файла %s: %v\n", file.Name(), err)
	}
}

func closeInputFile(file *os.File) {
	atomic.AddInt64(&openInputFiles, -1) //для подсчета открытых файлов
	closeFile(file)
}

func processingFiles(ctx context.Context, filePaths []string) ([]<-chan int, error) {
	g, _ := errgroup.WithContext(context.Background())
	channels := make([]<-chan int, len(filePaths))

	for i, filePath := range filePaths {
		g.Go(
			func() error {
				file, err := os.Open(filePath)
				if err != nil {
					return fmt.Errorf("ошибка открытия файла %s: %w", filePath, err)
				}
				atomic.AddInt64(&openInputFiles, 1)
				channels[i] = readNumbers(ctx, file)
				return nil
			},
		)
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return channels, nil
}

func readNumbers(ctx context.Context, file *os.File) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		defer closeInputFile(file)
		scanner := bufio.NewScanner(file)
		var linesProcessed int64
		for scanner.Scan() {
			time.Sleep(time.Second) //TODO: задержка для просмотра прогресса в терминале
			if ctx.Err() != nil {
				fmt.Printf("Контекст отменен в readNumbers для файла %s: %v\n", file.Name(), ctx.Err())
				updateProcessedLines(file.Name(), linesProcessed)
				return
			}
			atomic.AddInt64(&totalProcessedLines, 1)
			if num, err := strconv.Atoi(scanner.Text()); err == nil {
				out <- num
				linesProcessed++
			} else {
				otherValues(scanner.Text())
				atomic.AddInt64(&errorLines, 1)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("Ошибка чтения файла: %v\n", err)
		}
		updateProcessedLines(file.Name(), linesProcessed)
	}()
	return out
}

func updateProcessedLines(file string, count int64) {
	mu.Lock()
	processedLines[file] += count
	mu.Unlock()
}

func otherValues(val string) {
	mu.Lock()
	defer mu.Unlock()

	writer := bufio.NewWriter(otherFile)
	_, err := writer.WriteString(fmt.Sprintf("%s\n", val))
	if err != nil {
		fmt.Printf("Ошибка записи в other файл: %v\n", err)
	}
	err = writer.Flush()
	if err != nil {
		fmt.Printf("Ошибка сбрасывания данных из буфера в other файл: %v\n", err)
	}
}

func merge(ctx context.Context, inputs ...<-chan int) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)

		// слайс входных каналов и учет открытых из них
		current := make([]int, len(inputs))
		activeInputs := len(inputs)

		// считываем первое значение из каждого канала
		for i, ch := range inputs {
			current[i], activeInputs = readFromChannel(ctx, ch, activeInputs)
		}

		for activeInputs > 0 {
			select {
			case <-ctx.Done():
				// Контекст был отменен, завершаем работу
				fmt.Println("Слияние прервано из-за отмены контекста")
				return
			default:
				minVal := math.MaxInt // максимальное значение int
				minIdxChannel := -1

				for i, val := range current {
					if val != math.MaxInt && val < minVal {
						minVal = val
						minIdxChannel = i
					}
				}

				// Отправляем минимальное значение в выходной канал
				select {
				case <-ctx.Done():
					// Контекст был отменен во время отправки
					fmt.Println("Слияние прервано из-за отмены контекста при отправке значения")
					return
				default:
					select {
					case output <- minVal:
						// Значение успешно отправлено
					}
				}

				// Считываем следующее значение из канала, откуда взяли минимальное
				current[minIdxChannel], activeInputs = readFromChannel(ctx, inputs[minIdxChannel], activeInputs)
			}
		}
	}()
	return output
}

func readFromChannel(ctx context.Context, ch <-chan int, activeCount int) (int, int) {
	marker := func() (int, int) {
		return math.MaxInt, activeCount - 1 // math.MaxInt маркер закрытого канала
	}
	select {
	case <-ctx.Done():
		return marker()
	default:
		select {
		case val, ok := <-ch:
			if ok {
				return val, activeCount
			}
			return marker()
		}
	}
}
