package handler

import (
	"fmt"
	"strings"
	"telegram-bot/internal/service"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CommandHandlers Хендлеры команд
type CommandHandlers func(h *Handler, msg *tgbotapi.Message) string

func (h *Handler) HandleCommand(command string, msg *tgbotapi.Message) string {
	if fn, ok := commandHandlers[command]; ok {
		return fn(h, msg)
	}
	return "❌ Неизвестная команда"
}

// Маппинг команд на хендлеры
var commandHandlers = map[string]CommandHandlers{
	"start":    cmdStart,
	"help":     cmdHelp,
	"stats":    cmdStats,
	"history":  cmdHistory,
	"profile":  cmdProfile,
	"echo":     cmdEcho,
	"time":     cmdTime,
	"about":    cmdAbout,
	"register": cmdRegister,
	"cancel":   cmdCancel,
}

type Handler struct {
	userService     *service.UserService
	messageService  *service.MessageService
	registerService *service.RegisterService
}

func NewHandler(us *service.UserService, ms *service.MessageService, rs *service.RegisterService) *Handler {
	return &Handler{
		userService:     us,
		messageService:  ms,
		registerService: rs,
	}
}

func cmdStart(_ *Handler, _ *tgbotapi.Message) string {
	return `👋 <b>Добро пожаловать!</b>

<b>Доступные команды:</b>
/help - помощь
/stats - статистика
/history - история
/register - регистрация
/cancel - отмена`
}

func cmdHelp(_ *Handler, _ *tgbotapi.Message) string {
	return `📚 <b>Справка:</b>

/start - начало
/help - эта справка
/register - регистрация
/stats - статистика
/history - история
/profile - профиль
/echo [текст] - эхо
/time - время
/about - о боте`
}

func cmdStats(h *Handler, msg *tgbotapi.Message) string {
	userCount, msgCount, err := h.userService.GetStats()
	if err != nil {
		return "❌ Ошибка получения статистики"
	}
	return fmt.Sprintf("📊 <b>Статистика:</b>\n\n👥 Пользователей: %d\n💬 Сообщений: %d", userCount, msgCount)
}

func cmdHistory(h *Handler, msg *tgbotapi.Message) string {
	messages, err := h.messageService.GetHistory(msg.Chat.ID, 5)
	if err != nil {
		return "❌ Ошибка получения истории"
	}
	if len(messages) == 0 {
		return "📭 История пуста"
	}

	var result strings.Builder

	result.WriteString("📜 <b>История:</b>\n\n")
	for _, message := range messages {
		prefix := "👤"
		if !message.IsFromUser {
			prefix = "🤖"
		}
		result.WriteString(prefix + message.Text)
	}
	return result.String()
}

func cmdProfile(_ *Handler, msg *tgbotapi.Message) string {
	return fmt.Sprintf("👤 <b>Профиль:</b>\n\nID: <code>%d</code>\nUsername: @%s\nИмя: %s\nФамилия: %s",
		msg.Chat.ID, msg.From.UserName, msg.From.FirstName, msg.From.LastName)
}

func cmdEcho(_ *Handler, msg *tgbotapi.Message) string {
	if len(msg.CommandArguments()) > 0 {
		return msg.CommandArguments()
	}
	return "Укажите текст: /echo [текст]"
}

func cmdTime(_ *Handler, _ *tgbotapi.Message) string {
	return "🕒 " + time.Now().Format("2006-01-02 15:04:05")
}

func cmdAbout(_ *Handler, _ *tgbotapi.Message) string {
	return `🤖 <b>О боте</b>

Учебный проект на Go.
• SQLite
• Архитектура Clean Architecture`
}

func cmdRegister(_ *Handler, _ *tgbotapi.Message) string {
	// Заглушка - логика регистрации будет в Bot
	return ""
}

func cmdCancel(_ *Handler, _ *tgbotapi.Message) string {
	// Заглушка - логика отмены будет в Bot
	return ""
}
