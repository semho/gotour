package postgres

import (
	"chat/internal/models"
	"chat/pkg/customerrors"
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"time"
)

type Storage struct {
	db            *sql.DB
	maxChatSize   int
	maxChatsCount int
}

func NewPostgresStorage(dsn string, maxChatSize, maxChatsCount int) (*Storage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &Storage{
		db:            db,
		maxChatSize:   maxChatSize,
		maxChatsCount: maxChatsCount,
	}, nil
}

func (s *Storage) GetDefaultHistorySize() int {
	return s.maxChatSize
}

func (s *Storage) CreateSession(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (id, nickname)
		VALUES ($1, $2)`

	_, err := s.db.ExecContext(ctx, query, session.ID, session.Nickname)
	if err != nil {
		return customerrors.NewSessionError(session.ID, err)
	}
	return nil
}

func (s *Storage) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	query := `
		SELECT id, nickname
		FROM sessions
		WHERE id = $1`

	session := &models.Session{
		AnonNicknames: make(map[string]string),
	}

	err := s.db.QueryRowContext(ctx, query, sessionID).Scan(&session.ID, &session.Nickname)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, customerrors.NewSessionError(sessionID, customerrors.ErrSessionNotFound)
		}
		return nil, customerrors.NewSessionError(sessionID, err)
	}

	nicknamesQuery := `
		SELECT chat_id, nickname
		FROM anon_nicknames
		WHERE session_id = $1`

	rows, err := s.db.QueryContext(ctx, nicknamesQuery, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var chatID, nickname string
		if err := rows.Scan(&chatID, &nickname); err != nil {
			return nil, err
		}
		session.AnonNicknames[chatID] = nickname
	}

	return session, nil
}

func (s *Storage) CreateChat(ctx context.Context, chat *models.Chat) error {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM chats").Scan(&count)
	if err != nil {
		return customerrors.NewChatError(chat.ID, err)
	}
	if count >= s.maxChatsCount {
		return customerrors.NewChatError(chat.ID, customerrors.ErrMaxNumberReached)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO chats (id, history_size, ttl, read_only, private, owner_id)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = tx.ExecContext(ctx, query, chat.ID, chat.HistorySize, chat.TTL, chat.ReadOnly, chat.Private, chat.OwnerID)
	if err != nil {
		return customerrors.NewChatError(chat.ID, err)
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO anon_counts (chat_id, count) VALUES ($1, 0)", chat.ID)
	if err != nil {
		return customerrors.NewChatError(chat.ID, err)
	}

	return tx.Commit()
}

func (s *Storage) GetChat(ctx context.Context, chatID string) (*models.Chat, error) {
	query := `
		SELECT c.id, c.history_size, c.ttl, c.read_only, c.private, c.owner_id
		FROM chats c
		WHERE c.id = $1`

	chat := &models.Chat{
		Messages:     make([]models.Message, 0),
		AllowedUsers: make([]string, 0),
	}

	var ttl sql.NullTime
	err := s.db.QueryRowContext(ctx, query, chatID).Scan(
		&chat.ID,
		&chat.HistorySize,
		&ttl,
		&chat.ReadOnly,
		&chat.Private,
		&chat.OwnerID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, customerrors.NewChatError(chatID, customerrors.ErrChatNotFound)
		}
		return nil, customerrors.NewChatError(chatID, err)
	}

	if ttl.Valid {
		chat.TTL = &ttl.Time
	}

	allowedUsersQuery := `
		SELECT session_id
		FROM chat_access
		WHERE chat_id = $1 AND granted = true`

	rows, err := s.db.QueryContext(ctx, allowedUsersQuery, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sessionID string
		if err := rows.Scan(&sessionID); err != nil {
			return nil, err
		}
		chat.AllowedUsers = append(chat.AllowedUsers, sessionID)
	}

	return chat, nil
}

func (s *Storage) GetAndIncrementAnonCount(ctx context.Context, chatID string) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var count int
	err = tx.QueryRowContext(
		ctx,
		`UPDATE anon_counts 
		SET count = count + 1 
		WHERE chat_id = $1 
		RETURNING count`, chatID,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, tx.Commit()
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) DeleteChat(ctx context.Context, chatID string) error {
	query := `DELETE FROM chats WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, chatID)
	if err != nil {
		return customerrors.NewChatError(chatID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return customerrors.NewChatError(chatID, err)
	}

	if rows == 0 {
		return customerrors.NewChatError(chatID, customerrors.ErrChatNotFound)
	}

	return nil
}

func (s *Storage) SetChatTTL(ctx context.Context, chatID string, ttl time.Time) error {
	query := `
        UPDATE chats 
        SET ttl = $1 
        WHERE id = $2`

	result, err := s.db.ExecContext(ctx, query, ttl, chatID)
	if err != nil {
		return customerrors.NewChatError(chatID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return customerrors.NewChatError(chatID, err)
	}

	if rows == 0 {
		return customerrors.NewChatError(chatID, customerrors.ErrChatNotFound)
	}

	return nil
}

func (s *Storage) AddMessage(ctx context.Context, message *models.Message) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentSize int
	err = tx.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM messages WHERE chat_id = $1`,
		message.ChatID,
	).Scan(&currentSize)
	if err != nil {
		return customerrors.NewChatError(message.ChatID, err)
	}

	var historySize int
	err = tx.QueryRowContext(
		ctx,
		`SELECT history_size FROM chats WHERE id = $1`,
		message.ChatID,
	).Scan(&historySize)
	if err != nil {
		if err == sql.ErrNoRows {
			return customerrors.NewChatError(message.ChatID, customerrors.ErrChatNotFound)
		}
		return customerrors.NewChatError(message.ChatID, err)
	}

	if historySize > 0 && currentSize >= historySize {
		_, err = tx.ExecContext(
			ctx,
			`DELETE FROM messages 
            WHERE id IN (
                SELECT id FROM messages 
                WHERE chat_id = $1 
                ORDER BY timestamp ASC 
                LIMIT $2
            )`,
			message.ChatID,
			currentSize-historySize+1,
		)
		if err != nil {
			return customerrors.NewChatError(message.ChatID, err)
		}
	}

	query := `
        INSERT INTO messages (id, chat_id, session_id, nickname, text, timestamp)
        VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = tx.ExecContext(
		ctx, query,
		message.ID,
		message.ChatID,
		message.SessionID,
		message.Nickname,
		message.Text,
		message.Timestamp,
	)
	if err != nil {
		return customerrors.NewChatError(message.ChatID, err)
	}

	return tx.Commit()
}

func (s *Storage) GetChatHistory(ctx context.Context, chatID string) ([]*models.Message, error) {
	query := `
        SELECT id, chat_id, session_id, nickname, text, timestamp
        FROM messages
        WHERE chat_id = $1
        ORDER BY timestamp ASC`

	rows, err := s.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, customerrors.NewChatError(chatID, err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		msg := &models.Message{}
		err := rows.Scan(
			&msg.ID,
			&msg.ChatID,
			&msg.SessionID,
			&msg.Nickname,
			&msg.Text,
			&msg.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (s *Storage) RequestChatAccess(ctx context.Context, chatID, sessionID string) error {
	var private bool
	var ownerID string
	err := s.db.QueryRowContext(
		ctx,
		`SELECT private, owner_id FROM chats WHERE id = $1`,
		chatID,
	).Scan(&private, &ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return customerrors.NewChatError(chatID, customerrors.ErrChatNotFound)
		}
		return customerrors.NewChatError(chatID, err)
	}

	if ownerID == sessionID {
		return customerrors.NewChatError(chatID, customerrors.ErrAccessAlreadyExist)
	}

	var exists bool
	err = s.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM chat_access WHERE chat_id = $1 AND session_id = $2)`,
		chatID, sessionID,
	).Scan(&exists)
	if err != nil {
		return customerrors.NewChatError(chatID, err)
	}
	if exists {
		return customerrors.NewChatError(chatID, customerrors.ErrAccessAlreadyRequested)
	}

	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO chat_access (chat_id, session_id, granted)
         VALUES ($1, $2, false)`,
		chatID, sessionID,
	)
	if err != nil {
		return customerrors.NewChatError(chatID, err)
	}

	return nil
}

func (s *Storage) GetAccessRequests(ctx context.Context, chatID string) ([]string, error) {
	query := `
        SELECT session_id
        FROM chat_access
        WHERE chat_id = $1 AND granted = false`

	rows, err := s.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, customerrors.NewChatError(chatID, err)
	}
	defer rows.Close()

	var requests []string
	for rows.Next() {
		var sessionID string
		if err := rows.Scan(&sessionID); err != nil {
			return nil, err
		}
		requests = append(requests, sessionID)
	}

	return requests, nil
}

func (s *Storage) GrantChatAccess(ctx context.Context, chatID, sessionID string) error {
	query := `
        UPDATE chat_access
        SET granted = true
        WHERE chat_id = $1 AND session_id = $2`

	result, err := s.db.ExecContext(ctx, query, chatID, sessionID)
	if err != nil {
		return customerrors.NewChatError(chatID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return customerrors.NewChatError(chatID, err)
	}

	if rows == 0 {
		return customerrors.NewChatError(chatID, customerrors.ErrChatNotFound)
	}

	return nil
}

func (s *Storage) HasChatAccess(ctx context.Context, chatID, sessionID string) (bool, error) {
	var private bool
	var ownerID string
	err := s.db.QueryRowContext(
		ctx,
		`SELECT private, owner_id FROM chats WHERE id = $1`,
		chatID,
	).Scan(&private, &ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, customerrors.NewChatError(chatID, customerrors.ErrChatNotFound)
		}
		return false, customerrors.NewChatError(chatID, err)
	}

	if !private {
		return true, nil
	}

	if ownerID == sessionID {
		return true, nil
	}

	var granted bool
	err = s.db.QueryRowContext(
		ctx,
		`SELECT granted FROM chat_access WHERE chat_id = $1 AND session_id = $2`,
		chatID, sessionID,
	).Scan(&granted)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, customerrors.NewChatError(chatID, err)
	}

	return granted, nil
}

func (s *Storage) IsChatOwner(ctx context.Context, chatID, sessionID string) (bool, error) {
	var ownerID string
	err := s.db.QueryRowContext(
		ctx,
		`SELECT owner_id FROM chats WHERE id = $1`,
		chatID,
	).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, customerrors.NewChatError(chatID, customerrors.ErrChatNotFound)
		}
		return false, customerrors.NewChatError(chatID, err)
	}

	return ownerID == sessionID, nil
}
