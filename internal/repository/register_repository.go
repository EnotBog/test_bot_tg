package repository

import (
	"database/sql"
	"fmt"
	"telegram-bot/internal/models"
)

type RegisterRepository struct {
	db *sql.DB
}

func NewRegisterRepository(db *sql.DB) *RegisterRepository {
	return &RegisterRepository{db: db}
}

func (r *RegisterRepository) CreateOrUpdate(data *models.RegistrationData) error {
	// Проверяем существование
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users_registers WHERE chat_id = ?)`, data.ChatID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error while checking if user exists user_register: %v\n", err)
	}

	if exists {
		_, err = r.db.Exec(
			`UPDATE users_registers SET name = ?, email = ?, number_phone = ? WHERE chat_id = ?`,
			data.Name, data.Email, data.Phone, data.ChatID,
		)
	} else {
		_, err = r.db.Exec(
			`INSERT INTO users_registers (chat_id, name, email, number_phone) VALUES (?, ?, ?, ?)`,
			data.ChatID, data.Name, data.Email, data.Phone,
		)
	}
	return fmt.Errorf("error update user_register: %v\n", err)
}

func (r *RegisterRepository) GetByChatID(chatID int64) (*models.RegistrationData, error) {
	var data models.RegistrationData
	err := r.db.QueryRow(
		`SELECT chat_id, name, email, number_phone FROM users_registers WHERE chat_id = ?`,
		chatID,
	).Scan(&data.ChatID, &data.Name, &data.Email, &data.Phone)

	if err != nil {
		return nil, fmt.Errorf("error while querying user_register: %v\n", err)
	}
	return &data, nil
}
