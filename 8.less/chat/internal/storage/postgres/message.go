package postgres

import (
	"chat/internal/storage"
	kafka_v1 "chat/pkg/kafka/v1"
	"context"
	"database/sql"
)

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) storage.MessageRepository {
	return &postgresRepository{
		db: db,
	}
}

func (r *postgresRepository) GetChatHistorySize(ctx context.Context, chatID string) (int, error) {
	var historySize int
	err := r.db.QueryRowContext(
		ctx,
		`SELECT history_size FROM chats WHERE id = $1`,
		chatID,
	).Scan(&historySize)
	if err != nil {
		return 0, err
	}
	return historySize, nil
}

func (r *postgresRepository) GetCurrentMessagesCount(ctx context.Context, chatID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM messages WHERE chat_id = $1`,
		chatID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *postgresRepository) DeleteOldestMessages(ctx context.Context, chatID string, count int) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM messages 
         WHERE id IN (
             SELECT id FROM messages 
             WHERE chat_id = $1 
             ORDER BY timestamp ASC 
             LIMIT $2
         )`,
		chatID,
		count,
	)
	return err
}

func (r *postgresRepository) SaveMessage(ctx context.Context, event *kafka_v1.ChatMessageEvent) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO messages (id, chat_id, session_id, nickname, text, timestamp)
         VALUES ($1, $2, $3, $4, $5, $6)`,
		event.Payload.MessageId,
		event.Payload.ChatId,
		event.Payload.SessionId,
		event.Payload.Nickname,
		event.Payload.Text,
		event.Payload.Timestamp.AsTime(),
	)
	return err
}
