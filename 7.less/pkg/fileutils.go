package pkg

import (
	"bufio"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

var (
	openInputFiles int64
)

func CloseFile(file *os.File, log *logrus.Logger) {
	err := file.Close()
	if err != nil {
		log.Errorf("Ошибка при закрытии файла %s: %v", file.Name(), err)
	}
}

func CloseInputFile(file *os.File, log *logrus.Logger) {
	atomic.AddInt64(&openInputFiles, -1)
	CloseFile(file, log)
}

func OpenInputFile(filePath string) (*os.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	atomic.AddInt64(&openInputFiles, 1)
	return file, nil
}

func WriteToFile(file *os.File, content string) error {
	writer := bufio.NewWriter(file)
	_, err := writer.WriteString(fmt.Sprintf("%s\n", content))
	if err != nil {
		return err
	}
	return writer.Flush()
}

func GetOpenInputFilesCount() int64 {
	return atomic.LoadInt64(&openInputFiles)
}
