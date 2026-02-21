package main

import "time"

// User Структура для хранения пользователя
type User struct {
	Username      string    `json:"username"`
	LastName      string    `json:"last_name"`
	FirstName     string    `json:"first_name"`
	CreatedAt     time.Time `json:"created_at"`
	LastActive    time.Time `json:"last_active"`
	ChatID        int64     `json:"chat_id"`
	ID            int64     `json:"id"`
	CurrentAction string    `json:"last_action"`
	IsActive      bool      `json:"is_active"`
	SessionActive bool      `json:"session_active"`
}
