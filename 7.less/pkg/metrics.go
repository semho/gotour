package pkg

import (
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

var (
	processedLines      map[string]int64
	errorLines          int64
	totalProcessedLines int64
	mu                  sync.Mutex
)

func init() {
	processedLines = make(map[string]int64)
}

func IncrementTotalProcessedLines() {
	atomic.AddInt64(&totalProcessedLines, 1)
}

func IncrementErrorLines() {
	atomic.AddInt64(&errorLines, 1)
}

func UpdateProcessedLines(file string, count int64) {
	mu.Lock()
	defer mu.Unlock()
	processedLines[file] += count
}

func PrintMetrics(log *logrus.Logger, openInputFiles int64) {
	log.Infof(
		"Метрики: Обработано строк: %d, Строк с ошибками: %d, Открыто входных файлов: %d",
		atomic.LoadInt64(&totalProcessedLines), atomic.LoadInt64(&errorLines), openInputFiles,
	)
}

func PrintFinalCounters(log *logrus.Logger, openInputFiles int64) {
	log.Info("Финальные счетчики:")
	mu.Lock()
	for file, count := range processedLines {
		log.Infof("Файл %s: обработано строк %d", file, count)
	}
	mu.Unlock()
	log.Infof("Всего обработано строк: %d", atomic.LoadInt64(&totalProcessedLines))
	log.Infof("Всего строк с ошибками: %d", atomic.LoadInt64(&errorLines))
	log.Infof("Открытых входных файлов: %d", openInputFiles)
}
