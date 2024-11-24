package consumer

import (
	kafka_v1 "chat/pkg/kafka/v1"
	"chat/pkg/logger"
	"context"
	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"
)

type MessageHandler func(context.Context, *kafka_v1.ChatMessageEvent) error

type Consumer interface {
	Start(ctx context.Context, handler MessageHandler) error
	Close() error
}

type kafkaConsumer struct {
	consumer sarama.ConsumerGroup
	topic    string
}

func NewKafkaConsumer(brokers []string, groupID string, topic string) (Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1000

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &kafkaConsumer{
		consumer: consumer,
		topic:    topic,
	}, nil
}

func (k kafkaConsumer) Start(ctx context.Context, handler MessageHandler) error {
	groupHandler := &consumerGroupHandler{
		handler: handler,
		ctx:     ctx,
	}

	for {
		err := k.consumer.Consume(ctx, []string{k.topic}, groupHandler)
		if err != nil {
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (k kafkaConsumer) Close() error {
	return k.consumer.Close()
}

type consumerGroupHandler struct {
	handler MessageHandler
	ctx     context.Context
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	for {
		select {
		case msg := <-claim.Messages():
			if msg == nil {
				return nil
			}

			var event kafka_v1.ChatMessageEvent
			if err := proto.Unmarshal(msg.Value, &event); err != nil {
				logger.Log.Error(
					"Failed to unmarshal message: %v, topic: %s, partition: %d, offset: %d",
					err, msg.Topic, msg.Partition, msg.Offset,
				)
				session.MarkMessage(msg, "")
				continue
			}

			if err := h.handler(h.ctx, &event); err != nil {
				//TODO: выводить ошибку
				continue
			}

			//отмечаем сообщение прочитанным
			session.MarkMessage(msg, "")

		case <-h.ctx.Done():
			return nil
		}
	}
}
