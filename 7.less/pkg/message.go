package pkg

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	"time"
)

type Message struct {
	Type    MessageType
	Content string
}

type MessageType int

const (
	MessageTypeError MessageType = iota
	MessageTypeInfo
	MessageTypeWarn
	MessageTypeDebug
	MessageTypeContext
)

func (m Message) IsError() bool {
	return m.Type == MessageTypeError
}

func HandleMessages(ctx context.Context, msgChan <-chan Message, resultChan <-chan struct{}, log *logrus.Logger) {
	span := sentry.StartSpan(ctx, "handle_messages")
	defer span.Finish()

	for {
		select {
		case msg := <-msgChan:
			handleMessage(msg, log)
			if msg.IsError() {
				return // выходим после обработки критической ошибки
			}
		case <-resultChan:
			return
		case <-ctx.Done():
			for { //читаем все сообщения контекста
				select {
				case msg := <-msgChan:
					handleMessage(msg, log)
				case <-time.After(1000 * time.Millisecond):
					return
				}
			}
		}
	}
}

func handleMessage(msg Message, log *logrus.Logger) {
	switch msg.Type {
	case MessageTypeError:
		log.Errorf("Ошибка обработки файлов: %v", msg.Content)
		sentry.CaptureException(fmt.Errorf(msg.Content))
	case MessageTypeDebug:
		log.Debug(msg.Content)
		sentry.CaptureMessage(msg.Content)
	case MessageTypeWarn:
		log.Warn(msg.Content)
		sentry.CaptureMessage(msg.Content)
	case MessageTypeInfo:
		log.Info(msg.Content)
	case MessageTypeContext:
		log.Warn(msg.Content)
		sentry.CaptureMessage(msg.Content)
	}
}
