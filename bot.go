package main

import (
	"database/sql"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	regMu       sync.RWMutex
	mesMu       sync.RWMutex
	api         *tgbotapi.BotAPI
	db          *sql.DB
	regState    map[int64]*RegistrationData
	activeUsers map[int64]*User
}

func (bot *Bot) getRegister() map[int64]*RegistrationData {
	return make(map[int64]*RegistrationData)
}

func (bot *Bot) clearActiveUsers() bool {
	return true
}
