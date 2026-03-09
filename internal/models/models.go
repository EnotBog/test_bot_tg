package models

import "time"

type User struct {
	ID            int64     `json:"id"`
	ChatID        int64     `json:"chat_id"`
	Username      string    `json:"username"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	CreatedAt     time.Time `json:"created_at"`
	LastActive    time.Time `json:"last_active"`
	IsActive      bool      `json:"is_active"`
}

type RegistrationData struct {
	Step   RegistrationStep `json:"step"`
	ChatID int64             `json:"chat_id"`
	Name   string            `json:"name"`
	Email  string            `json:"email"`
	Phone  string            `json:"phone"`
}

type RegistrationStep int

const (
	StepNone RegistrationStep = iota
	StepCheck
	StepName
	StepEmail
	StepPhone
	StepComplete
	StepCorrect
)

type Session struct {
	User          *User
	CurrentAction string
	Active        bool
}

type Message struct {
	ID         int64     `json:"id"`
	ChatID     int64     `json:"chat_id"`
	Text       string    `json:"message_text"`
	IsFromUser bool      `json:"is_from_user"`
	IsCommand  bool      `json:"is_command"`
	CreatedAt  time.Time `json:"created_at"`
}

type Stats struct {
	UserCount    int `json:"user_count"`
	MessageCount int `json:"message_count"`
}
