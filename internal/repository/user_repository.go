package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"telegram-bot/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) CreateOrUpdate(chatID int64, username, firstName, lastName string) (int64, error) {
	var id int64

	err := ur.db.QueryRow(`SELECT id FROM users WHERE chat_id = ?`, chatID).Scan(&id)
	if err == nil {
		_, err = ur.db.Exec(
			`UPDATE users 
					SET username = ?, first_name = ?, last_name = ?, last_active = ? 
					WHERE chat_id = ?`,
			username, firstName, lastName, time.Now(), chatID,
		)
		log.Printf("Пользователь обновлен: ChatID=%d, ID в БД=%d", chatID, id)
		return id, err
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("ошибка проверки пользователя: %v", err)
	}

	res, err := ur.db.Exec(
		`INSERT INTO users (chat_id, username, first_name, last_name, created_at, last_active)
		VALUES (?, ?, ?, ?, ?, ?)`,
		chatID, username, firstName, lastName, time.Now(), time.Now(),
	)
	if err != nil {
		return 0, err
	}
	log.Printf("Add new user, ChatID:%d", chatID)
	return res.LastInsertId()
}

func (ur *UserRepository) GetByChatID(chatID int64) (*models.User, error) {
	var user models.User
	err := ur.db.QueryRow(
		`SELECT id, chat_id, username, first_name, last_name, created_at, last_active, is_active 
		FROM users WHERE chat_id = ?`,
		chatID,
	).Scan(&user.ID, &user.ChatID, &user.Username, &user.FirstName, &user.LastName, &user.CreatedAt, &user.LastActive, &user.IsActive)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) GetStats() (int, int, error) {
	var userCount, messageCount int
	err := ur.db.QueryRow(`SELECT count(*) FROM users`).Scan(&userCount)
	if err != nil {
		return 0, 0, fmt.Errorf("error getting stats user count: %v\n", err)
	}
	err = ur.db.QueryRow(`SELECT count(*) FROM messages`).Scan(&messageCount)
	if err != nil {
		return 0, 0, fmt.Errorf("error getting stats message count: %v\n", err)
	}
	return userCount, messageCount, nil
}
