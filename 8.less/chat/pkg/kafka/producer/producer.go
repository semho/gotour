package producer

import (
	kafka_v1 "chat/pkg/kafka/v1"
	"context"

	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"
)

type Producer interface {
	SendMessage(ctx context.Context, event *kafka_v1.ChatMessageEvent) error
	Close() error
}

type kafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewKafkaProducer(brokers []string, topic string) (Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Retry.Max = 5
	config.Producer.Retry.Backoff = 100

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &kafkaProducer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *kafkaProducer) SendMessage(_ context.Context, event *kafka_v1.ChatMessageEvent) error {
	if err := event.ValidateAll(); err != nil {
		return err
	}

	data, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(event.Payload.ChatId),
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event_type"),
				Value: []byte(event.Metadata.EventType.String()),
			},
			{
				Key:   []byte("event_id"),
				Value: []byte(event.Metadata.EventId),
			},
			{
				Key:   []byte("message_id"),
				Value: []byte(event.Payload.MessageId),
			},
		},
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}

func (p *kafkaProducer) Close() error {
	return p.producer.Close()
}
