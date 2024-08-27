package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
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

	log *logrus.Logger
)

func init() {
	var err error
	otherFile, err = os.Create("other")
	if err != nil {
		log.Fatal("Ошибка открытия файла 'other':", err)
	}
	processedLines = make(map[string]int64)

	// Инит logrus
	log = logrus.New()
	log.SetFormatter(
		&logrus.TextFormatter{
			FullTimestamp: true,
		},
	)
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

	// Инит Sentry
	err = sentry.Init(
		sentry.ClientOptions{
			Dsn:              "https://27cdf21a098c379cef0a9ffb9259e6b1@o4507807833849856.ingest.de.sentry.io/4507807837651024",
			EnableTracing:    true,
			TracesSampleRate: 1.0, //трасировка 100%
			//Debug:            true,
		},
	)
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
}

func main() {
	defer closeFile(otherFile)
	defer sentry.Flush(2 * time.Second)

	var inputFiles arrayFlags
	var outputFile string
	var logLevel string
	flag.Var(&inputFiles, "inputs", "файлы для чтения через запятую")
	flag.StringVar(&outputFile, "output", "output", "название выходного файла")
	flag.StringVar(&logLevel, "log-level", "info", "уровень логирования (debug, info, warn, error)")
	flag.Parse()

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		log.Warnf("Неверный уровень логирования: %s. Используется уровень по умолчанию (info)", logLevel)
	} else {
		log.SetLevel(level)
	}

	log.Info("Запуск программы")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	metricsCtx, metricsCancel := context.WithCancel(ctx)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go reportMetrics(metricsCtx, wg)

	transaction := sentry.StartTransaction(ctx, "process_files")
	defer transaction.Finish()

	if err = processFiles(transaction.Context(), inputFiles, outputFile); err != nil {
		log.Errorf("Ошибка обработки файлов: %v", err)
		sentry.CaptureException(err)
	}

	metricsCancel()
	wg.Wait()
	printFinalCounters()

	log.Info("Программа завершена")
}

func reportMetrics(ctx context.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(1 * time.Second) //как часто смотрим метрики
	defer ticker.Stop()
	defer wg.Done()
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
	log.Infof(
		"Метрики: Обработано строк: %d, Строк с ошибками: %d, Открыто входных файлов: %d",
		atomic.LoadInt64(&totalProcessedLines), atomic.LoadInt64(&errorLines), atomic.LoadInt64(&openInputFiles),
	)
}

func printFinalCounters() {
	log.Info("Финальные счетчики:")
	mu.Lock()
	for file, count := range processedLines {
		log.Infof("Файл %s: обработано строк %d", file, count)
	}
	mu.Unlock()
	log.Infof("Всего обработано строк: %d", atomic.LoadInt64(&totalProcessedLines))
	log.Infof("Всего строк с ошибками: %d", atomic.LoadInt64(&errorLines))
	log.Infof("Открытых входных файлов: %d", atomic.LoadInt64(&openInputFiles))
}

func processFiles(ctx context.Context, inputFiles []string, outputFile string) error {
	log.Info("Начало обработки файлов")
	span := sentry.StartSpan(ctx, "process_files")
	defer span.Finish()

	errChan := make(chan error, 1)
	var err error

	channels := processInputFiles(span.Context(), inputFiles, errChan)
	select {
	case err = <-errChan:
		return err
	default:
	}

	mergeSpan := sentry.StartSpan(span.Context(), "merge_channels")
	mergedChannel := merge(mergeSpan.Context(), channels...)
	mergeSpan.Finish()

	saveSpan := sentry.StartSpan(span.Context(), "save_result")
	err = saveResult(saveSpan.Context(), outputFile, mergedChannel)
	saveSpan.Finish()

	return err
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Errorf("Ошибка при закрытии файла %s: %v", file.Name(), err)
	}
}

func closeInputFile(file *os.File) {
	atomic.AddInt64(&openInputFiles, -1) //для подсчета открытых файлов
	closeFile(file)
}

func processInputFiles(ctx context.Context, filePaths []string, errChan chan<- error) []<-chan int {
	done := make(chan struct{})
	var once sync.Once
	closeDone := func() {
		once.Do(
			func() {
				close(done)
			},
		)
	}
	defer closeDone()

	span := sentry.StartSpan(ctx, "process_input_files")
	defer span.Finish()

	channels := make([]<-chan int, len(filePaths))

	wg := &sync.WaitGroup{}

	for i, filePath := range filePaths {
		wg.Add(1)
		go func(i int, filePath string) {
			defer wg.Done()
			fileSpan := sentry.StartSpan(span.Context(), "process_file")
			fileSpan.SetTag("file", filePath)
			defer fileSpan.Finish()

			//TODO: без задержки не увидим сообщение о закрытии контекста
			time.Sleep(10 * time.Millisecond)
			select {
			case <-ctx.Done():
				sentry.CaptureMessage(
					fmt.Sprintf(
						"Закрытие по контексту в processInputFiles %s: %v",
						filePath,
						ctx.Err(),
					),
				)
				return
			case <-done: //нужно для завершения работы, если есть ошибка
				return
			default:
				file, err := os.Open(filePath)
				if err != nil {
					sentry.CaptureException(err)
					errChan <- fmt.Errorf("В processInputFiles  ошибка открытия файла %s: %w", filePath, err)
					closeDone()
					return
				}
				atomic.AddInt64(&openInputFiles, 1)
				channels[i] = readNumbers(fileSpan.Context(), file, errChan)
			}
		}(i, filePath)
	}
	wg.Wait()

	return channels
}

func readNumbers(ctx context.Context, file *os.File, errChan chan<- error) <-chan int {
	span := sentry.StartSpan(ctx, "read_numbers")
	defer span.Finish()

	out := make(chan int)
	go func() {
		defer close(out)
		defer closeInputFile(file)
		scanner := bufio.NewScanner(file)
		var linesProcessed int64
		for scanner.Scan() {
			time.Sleep(time.Second) //TODO: задержка для просмотра прогресса в терминале
			if ctx.Err() != nil {
				log.Warnf("Контекст отменен в readNumbers для файла %s: %v", file.Name(), ctx.Err())
				sentry.CaptureMessage(
					fmt.Sprintf(
						"Контекст отменен в readNumbers для файла %s: %v",
						file.Name(),
						ctx.Err(),
					),
				)
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
				log.Infof("Пропущена некорректная строка в файле %s: %s", file.Name(), scanner.Text())
				sentry.CaptureMessage(
					fmt.Sprintf(
						"Пропущена некорректная строка в файле %s: %s",
						file.Name(),
						scanner.Text(),
					),
				)
			}
		}
		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("ошибка чтения файла %s: %v", file.Name(), err)
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
		log.Errorf("Ошибка записи в other файл: %v", err)
	}
	err = writer.Flush()
	if err != nil {
		log.Errorf("Ошибка сбрасывания данных из буфера в other файл: %v", err)
	}
}

func merge(ctx context.Context, inputs ...<-chan int) <-chan int {
	span := sentry.StartSpan(ctx, "merge")
	defer span.Finish()

	output := make(chan int)
	go func() {
		defer close(output)

		// слайс входных каналов и учет открытых из них
		current := make([]int, len(inputs))
		activeInputs := len(inputs)

		// считываем первое значение из каждого канала
		for i, ch := range inputs {
			current[i], activeInputs = readFromChannel(span.Context(), ch, activeInputs)
		}

		for activeInputs > 0 {
			select {
			case <-span.Context().Done():
				// Контекст был отменен, завершаем работу
				log.Warnf("Слияние прервано из-за отмены контекста %v", span.Context().Err())
				sentry.CaptureMessage("Слияние прервано из-за отмены контекста")
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
				case <-span.Context().Done():
					// Контекст был отменен во время отправки
					log.Warnf("Слияние прервано из-за отмены контекста при отправке значения %v", span.Context().Err())
					sentry.CaptureMessage("Слияние прервано из-за отмены контекста при отправке значения")
					return
				default:
					select {
					case output <- minVal:
						// Значение успешно отправлено
					}
				}

				// Считываем следующее значение из канала, откуда взяли минимальное
				current[minIdxChannel], activeInputs = readFromChannel(
					span.Context(),
					inputs[minIdxChannel],
					activeInputs,
				)
			}
		}
	}()
	return output
}

func readFromChannel(ctx context.Context, ch <-chan int, activeCount int) (int, int) {
	span := sentry.StartSpan(ctx, "save_result")
	defer span.Finish()

	marker := func() (int, int) {
		return math.MaxInt, activeCount - 1 // math.MaxInt маркер закрытого канала
	}
	select {
	case <-ctx.Done():
		log.Warnf("Слияние прервано из-за отмены контекста при отправке значения %v", span.Context().Err())
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

func saveResult(ctx context.Context, outputFile string, mergedChannel <-chan int) error {
	span := sentry.StartSpan(ctx, "save_result")
	defer span.Finish()

	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Errorf("Ошибка создания выходного файла: %v", err)
		sentry.CaptureException(err)
		return fmt.Errorf("ошибка создания выходного файла: %w", err)
	}

	defer closeFile(outFile)

	writer := bufio.NewWriter(outFile)

	for {
		select {
		case <-span.Context().Done():
			log.Warn("Получен сигнал завершения. Сохранение промежуточных результатов...")
			sentry.CaptureMessage("Сохранение результатов прервано из-за отмены контекста")
			return nil
		default:
			select {
			case num, ok := <-mergedChannel:
				if !ok {
					log.Info("Завершение записи результатов")
					return writer.Flush()
				}
				_, err = writer.WriteString(fmt.Sprintf("%d\n", num))
				if err != nil {
					log.Errorf("Ошибка записи в файл: %v", err)
					sentry.CaptureException(err)
					return fmt.Errorf("ошибка записи в файл: %w", err)
				}
				err = writer.Flush()
				if err != nil {
					log.Errorf("Ошибка сбрасывания данных из буфера в выходной файл: %v", err)
					sentry.CaptureException(err)
					return fmt.Errorf("ошибка сбрасывания данных из буфера в выходной файл: %w", err)
				}
			}
		}
	}
}
