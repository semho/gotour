package main

import (
	"7less/pkg"
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
	log       *logrus.Logger
)

func init() {
	var err error
	otherFile, err = os.Create("other")
	if err != nil {
		log.Fatal("Ошибка открытия файла 'other':", err)
	}

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
	defer pkg.CloseFile(otherFile, log)
	defer sentry.Flush(2 * time.Second)

	var inputFiles arrayFlags
	var outputFile string
	var logLevel string
	flag.Var(&inputFiles, "inputs", "файлы для чтения через запятую")
	flag.StringVar(&outputFile, "output", "output", "название выходного файла")
	flag.StringVar(&logLevel, "log-level", "info", "уровень логирования (debug, info, warn, error)")
	flag.Parse()

	setLogLevel(logLevel)

	log.Info("Запуск программы")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	metricsCtx, metricsCancel := context.WithCancel(ctx)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go reportMetrics(metricsCtx, wg)

	transaction := sentry.StartTransaction(ctx, "process_files")
	defer transaction.Finish()

	processFiles(transaction.Context(), inputFiles, outputFile)

	metricsCancel()
	wg.Wait()
	printFinalCounters()

	log.Info("Программа завершена")
}

func setLogLevel(logLevel string) {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		log.Warnf("Неверный уровень логирования: %s. Используется уровень по умолчанию (info)", logLevel)
	} else {
		log.SetLevel(level)
	}
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
	pkg.PrintMetrics(log, pkg.GetOpenInputFilesCount())
}

func printFinalCounters() {
	pkg.PrintFinalCounters(log, pkg.GetOpenInputFilesCount())
}

func processFiles(ctx context.Context, inputFiles []string, outputFile string) {
	log.Info("Начало обработки файлов")
	span := sentry.StartSpan(ctx, "process_files")
	defer span.Finish()

	messageChan := make(chan pkg.Message, 100)
	resultChan := make(chan struct{})

	go func() {
		defer close(resultChan)
		channels := processInputFiles(ctx, inputFiles, messageChan)
		if channels == nil {
			return // выходим после первой ошибки, для предотвращения чтения из других файлов
		}
		mergedChannel := merge(ctx, messageChan, channels...)
		saveResult(ctx, outputFile, mergedChannel, messageChan)
	}()

	pkg.HandleMessages(span.Context(), messageChan, resultChan, log)
}

func processInputFiles(
	ctx context.Context,
	filePaths []string,
	msgChan chan<- pkg.Message,
) []<-chan int {
	span := sentry.StartSpan(ctx, "process_input_files")
	defer span.Finish()

	channels := make([]<-chan int, 0, len(filePaths))

	for _, filePath := range filePaths {
		select {
		case <-ctx.Done():
			return nil
		default:
			file, err := pkg.OpenInputFile(filePath)
			if err != nil {
				msgChan <- pkg.Message{
					Type:    pkg.MessageTypeError,
					Content: fmt.Sprintf("В processInputFiles ошибка открытия файла %s: %v", filePath, err),
				}
				return nil
			}
			channels = append(channels, readNumbers(span.Context(), file, msgChan))
		}
	}

	return channels
}

func readNumbers(ctx context.Context, file *os.File, msgChan chan<- pkg.Message) <-chan int {
	span := sentry.StartSpan(ctx, "read_numbers")
	defer span.Finish()

	out := make(chan int)
	go func() {
		defer close(out)
		defer pkg.CloseInputFile(file, log)
		scanner := bufio.NewScanner(file)
		var linesProcessed int64
		for scanner.Scan() {
			time.Sleep(time.Second) //TODO: задержка для просмотра прогресса в терминале
			if ctx.Err() != nil {
				msgChan <- pkg.Message{
					Type: pkg.MessageTypeContext,
					Content: fmt.Sprintf(
						"Контекст отменен в readNumbers при чтении файла %s: %v",
						file.Name(),
						ctx.Err(),
					),
				}
				updateProcessedLines(file.Name(), linesProcessed)
				return
			}
			pkg.IncrementTotalProcessedLines()
			if num, err := strconv.Atoi(scanner.Text()); err == nil {
				out <- num
				linesProcessed++
			} else {
				otherValues(scanner.Text())
				pkg.IncrementErrorLines()
				msgChan <- pkg.Message{
					Type:    pkg.MessageTypeInfo,
					Content: fmt.Sprintf("Пропущена некорректная строка в файле %s: %s", file.Name(), scanner.Text()),
				}
			}
		}
		if err := scanner.Err(); err != nil {
			msgChan <- pkg.Message{
				Type:    pkg.MessageTypeError,
				Content: fmt.Sprintf("ошибка чтения файла %s: %v", file.Name(), err),
			}
		}
		updateProcessedLines(file.Name(), linesProcessed)
	}()
	return out
}

func updateProcessedLines(file string, count int64) {
	pkg.UpdateProcessedLines(file, count)
}

func otherValues(val string) {
	err := pkg.WriteToFile(otherFile, val)
	if err != nil {
		log.Errorf("Ошибка записи в other файл: %v", err)
	}
}

func merge(ctx context.Context, msgChan chan<- pkg.Message, inputs ...<-chan int) <-chan int {
	span := sentry.StartSpan(ctx, "merge_channels")
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
			case <-ctx.Done():
				// Контекст был отменен, завершаем работу
				msgChan <- pkg.Message{
					Type: pkg.MessageTypeContext,
					Content: fmt.Sprintf(
						"Слияние прервано из-за отмены контекста %v", span.Context().Err(),
					),
				}
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
					msgChan <- pkg.Message{
						Type: pkg.MessageTypeContext,
						Content: fmt.Sprintf(
							"Слияние прервано из-за отмены контекста при отправке значения %v", span.Context().Err(),
						),
					}
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

func saveResult(ctx context.Context, outputFile string, mergedChannel <-chan int, msgChan chan<- pkg.Message) {
	span := sentry.StartSpan(ctx, "save_result")
	defer span.Finish()

	outFile, err := os.Create(outputFile)
	if err != nil {
		msgChan <- pkg.Message{
			Type:    pkg.MessageTypeError,
			Content: fmt.Sprint("ошибка создания выходного файла"),
		}
		return
	}

	defer pkg.CloseFile(outFile, log)

	for {
		select {
		case <-span.Context().Done():
			msgChan <- pkg.Message{
				Type: pkg.MessageTypeContext,
				Content: fmt.Sprintf(
					"Сохранение результатов прервано из-за отмены контекста %v", span.Context().Err(),
				),
			}
			return
		default:
			select {
			case num, ok := <-mergedChannel:
				if !ok {
					msgChan <- pkg.Message{
						Type:    pkg.MessageTypeInfo,
						Content: fmt.Sprint("Завершение записи результатов"),
					}
					return
				}
				if err = pkg.WriteToFile(outFile, fmt.Sprintf("%d", num)); err != nil {
					msgChan <- pkg.Message{
						Type:    pkg.MessageTypeError,
						Content: fmt.Sprintf("ошибка записи в файл: %v", err),
					}
					return
				}
			}
		}
	}
}
