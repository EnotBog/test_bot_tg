package repository

import (
	"database/sql"
	"fmt"
	"time"

	"telegram-bot/internal/models"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Save(chatID int64, text string, isFromUser, isCommand bool) error {
	_, err := r.db.Exec(
		`INSERT INTO messages (chat_id, message_text, is_from_user, is_command, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		chatID, text, isFromUser, isCommand, time.Now(),
	)
	return err
}

func (r *MessageRepository) GetByChatID(chatID int64, limit int) ([]models.Message, error) {
	rows, err := r.db.Query(
		`SELECT id, chat_id, message_text, is_from_user, is_command, created_at
		FROM messages WHERE chat_id = ? ORDER BY created_at DESC LIMIT ?`,
		chatID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.ChatID, &m.Text, &m.IsFromUser, &m.IsCommand, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("row scan get messages: %w\n", err)
		}
		messages = append(messages, m)
	}
	return messages, nil
}
