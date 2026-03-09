package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type commandHandler func(bot *Bot, msg *tgbotapi.Message, user *User) string

var commandHandlers = map[string]commandHandler{
	"start":    cmdStart,
	"help":     cmdHelp,     // Помощь в командах
	"stats":    cmdStats,    // Статистика бота
	"history":  cmdHistory,  // История сообщений пользователя
	"profile":  cmdProfile,  // Профиль пользователя
	"echo":     cmdEcho,     // повторение сообщения
	"time":     cmdTime,     // Время
	"about":    cmdAbout,    // информация о боте
	"register": cmdRegister, // Регистрация пользователя
	"cancel":   cmdCancel,   // Отмена всех активных диалогов и действий
}

func pathCommand(bot *Bot, msg *tgbotapi.Message, user *User, commandUser string) string {
	if command, ok := commandHandlers[commandUser]; ok {
		if commandUser == "register" {
			command(bot, msg, user)
		}
		return command(bot, msg, user)
	}
	return fmt.Sprintf("❓ Неизвестная команда. Используйте /help для списка команд")
}

func cmdStart(_ *Bot, _ *tgbotapi.Message, _ *User) string {
	return `👋 <b>Добро пожаловать!</b>

Я бот, написанный на Go, с интеграцией базы данных SQLite.

<b>Доступные команды:</b>
/help - показать помощь
/stats - статистика бота
/history - история сообщений
/about - информация о боте
/echo [текст] - повторить текст
/register - Регистрация пользователя
/cancel - Отмена текущей операции`
}

func cmdHelp(_ *Bot, _ *tgbotapi.Message, _ *User) string {
	return `📚 <b>Справка по командам:</b>

<b>Основные команды:</b>
/start - начать работу
/help - эта справка
/about - информация о боте
/register - Регистрация пользователя
/cancel - Отмена текущей операции


<b>Информационные команды:</b>
/stats - статистика бота
/history - последние сообщения
/profile - информация о вас

<b>Утилиты:</b>
/echo [текст] - повторить текст
/time - текущее время`
}

func cmdStats(bot *Bot, _ *tgbotapi.Message, _ *User) string {
	userCount, messageCount, err := getStats(bot.db)
	if err != nil {
		return "❌ Ошибка получения статистики"
	}
	return fmt.Sprintf(`📊 <b>Статистика бота:</b>

👥 Пользователей: <b>%d</b>
💬 Сообщений: <b>%d</b>`, userCount, messageCount)
}

func cmdHistory(bot *Bot, msg *tgbotapi.Message, _ *User) string {
	userID, _ := addOrUpdateUser(bot.db, msg.Chat.ID,
		msg.From.UserName, msg.From.FirstName, msg.From.LastName)

	messages, err := getUserMessages(bot.db, userID, 5)
	if err != nil {
		return "❌ Ошибка получения истории"
	}

	if len(messages) == 0 {
		return "📭 У вас еще нет сообщений в истории"
	}

	return "📜 <b>Последние сообщения:</b>\n\n" +
		strings.Join(messages, "\n")
}

func cmdProfile(_ *Bot, msg *tgbotapi.Message, _ *User) string {
	return fmt.Sprintf(`👤 <b>Ваш профиль:</b>

ID: <code>%d</code>
Username: @%s
Имя: %s
Фамилия: %s`,
		msg.Chat.ID,
		msg.From.UserName,
		msg.From.FirstName,
		msg.From.LastName)
}

func cmdEcho(_ *Bot, msg *tgbotapi.Message, _ *User) string {

	if len(msg.CommandArguments()) > 0 {
		return msg.CommandArguments()
	}
	return "Пожалуйста, укажите текст после команды /echo"
}

func cmdTime(_ *Bot, _ *tgbotapi.Message, _ *User) string {
	return "🕒 Текущее время: " + time.Now().Format("2006-01-02 15:04:05")
}

func cmdAbout(_ *Bot, _ *tgbotapi.Message, _ *User) string {
	return `🤖 <b>О боте</b>

Это учебный проект Telegram-бота на языке Go.

<b>Функциональность:</b>
• Работа с командами
• Интеграция с SQLite базой данных
• Сохранение пользователей и сообщений
• Статистика и история

<b>Технологии:</b>
• Go 1.21+
• SQLite
• go-telegram-bot-api`
}

func cmdRegister(bot *Bot, _ *tgbotapi.Message, user *User) string {
	fmt.Println("Вход в команду регистрации")

	err := bot.startRegistration(user)
	if err != nil {
		log.Println(err)
		return fmt.Sprintf("Ошибка выполнения команды регистрации. %s", err)
	}

	//sendInlineKeyboard(msg, "Выберите поле для заполнения!")
	/*
		Здесь вызывается функция создания кнопок и отправки
		предполагается сделать три кнопки для выбора, что заполнить
	*/

	// Здесь лучше обрабатывать общий случай
	return ""
}

func cmdCancel(bot *Bot, _ *tgbotapi.Message, user *User) string {
	bot.mesMu.Lock()
	action := user.CurrentAction
	bot.mesMu.Unlock()

	switch action {
	case "register":
		bot.clearRegistration(user.ChatID)
	default:
		return "❌ Нет активных операций!"
	}
	fmt.Printf("Пользователь:%+v\nотменяет действие:%v\n ", user, action)
	return fmt.Sprintf("❌ Операция: %v отменена\n", action)
}
