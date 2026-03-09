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

func (r *MessageRepository) Save(userID int64, text string, isFromUser, isCommand bool) error {
	_, err := r.db.Exec(
		`INSERT INTO messages (user_id, message_text, is_from_user, is_command, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		userID, text, isFromUser, isCommand, time.Now(),
	)
	return err
}

func (r *MessageRepository) GetByUserID(userID int64, limit int) ([]models.Message, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, message_text, is_from_user, is_command, created_at
		FROM messages WHERE user_id = ? ORDER BY created_at DESC LIMIT ?`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.UserID, &m.Text, &m.IsFromUser, &m.IsCommand, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("row scan get messages: %w\n", err)
		}
		messages = append(messages, m)
	}
	return messages, nil
}
