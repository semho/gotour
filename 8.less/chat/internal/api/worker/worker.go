package worker

import (
	"chat/internal/service"
	"chat/internal/storage/postgres"
	"chat/pkg/kafka/consumer"
	"context"
	"database/sql"
)

type MessageWorker struct {
	consumer consumer.Consumer
	service  service.MessageService
}

func NewMessageWorker(brokers []string, groupID, topic, dbDSN string) (*MessageWorker, error) {
	cons, err := consumer.NewKafkaConsumer(brokers, groupID, topic)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	repo := postgres.NewPostgresRepository(db)
	svc := service.NewMessageService(repo)

	return &MessageWorker{
		consumer: cons,
		service:  svc,
	}, nil
}

func (w *MessageWorker) Start(ctx context.Context) error {
	return w.consumer.Start(ctx, w.service.ProcessMessage)
}

func (w *MessageWorker) Close() error {
	return w.consumer.Close()
}
