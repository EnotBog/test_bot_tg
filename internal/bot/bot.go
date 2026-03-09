package bot

import (
	"log"
	"sync"

	"telegram-bot/internal/config"
	"telegram-bot/internal/handler"
	"telegram-bot/internal/models"
	"telegram-bot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api            *tgbotapi.BotAPI
	config         *config.Config
	handler        *handler.Handler
	userService    *service.UserService
	messageService *service.MessageService

	regState    map[int64]*models.RegistrationData
	activeUsers map[int64]*models.Session
	mu          sync.RWMutex
}

func NewBot(api *tgbotapi.BotAPI, cfg *config.Config, h *handler.Handler, us *service.UserService, ms *service.MessageService) *Bot {
	return &Bot{
		api:            api,
		config:         cfg,
		handler:        h,
		userService:    us,
		messageService: ms,
		regState:       make(map[int64]*models.RegistrationData),
		activeUsers:    make(map[int64]*models.Session),
	}
}

func (b *Bot) Start() {
	log.Printf("Бот запущен: @%s", b.api.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := b.api.GetUpdatesChan(updateConfig)

	for update := range updates {
		b.handleUpdate(update)
	}
}

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		go b.handleCallback(update.CallbackQuery)
		return
	}
	if update.Message != nil && update.Message.Text != "" {
		go b.handleMessage(update.Message)
	}
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	log.Printf("[%d] @%s: %s \n", chatID, msg.From.UserName, msg.Text)
	// Создаем или обновляем пользователя
	user, err := b.userService.GetOrCreate(chatID, msg.From.UserName, msg.From.FirstName, msg.From.LastName)
	if err != nil {
		log.Printf("Ошибка создания пользователя: %v", err)
	}
	// Сохраняем сообщение
	b.messageService.Save(chatID, msg.Text, true, msg.IsCommand())

	// Проверяем сессию
	session := b.getSession(chatID)
	if session.Active && session.CurrentAction != "" {
		b.handleSessionAction(msg, session)
		return
	}

	// Обрабатываем команду
	if msg.IsCommand() {
		b.handleCommand(msg, user)
		return
	}
	// Обычное сообщение
	b.sendMessage(chatID, "Напишите /help для списка команд")
}

func (b *Bot) handleCommand(msg *tgbotapi.Message, user *models.User) {
	command := msg.Command()
	response := b.handler.HandleCommand(command, msg)
	if response != "" {
		b.sendMessage(msg.Chat.ID, response)
	} else {
		b.sendMessage(msg.Chat.ID, "❌ Неизвестная команда. Используйте /help")
	}
}

func (b *Bot) handleCallback(callback *tgbotapi.CallbackQuery) {
	// Обработка inline кнопок
	b.api.Send(tgbotapi.NewCallback(callback.ID, ""))
}

func (b *Bot) handleSessionAction(msg *tgbotapi.Message, session *models.Session) {
	// Обработка активных сессий (регистрация и т.д.)
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Ошибка отправки: %v", err)
	}
}

func (b *Bot) getSession(chatID int64) *models.Session {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if session, ok := b.activeUsers[chatID]; ok {
		return session
	}

	// Возвращаем новую пустую сессию вместо nil
	return &models.Session{}
}
