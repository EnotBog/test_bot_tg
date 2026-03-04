package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func (bot *Bot) handleIncomingMessage(update tgbotapi.Update) {
	msg := update.Message
	var user *User
	var response string

	// Логируем сообщение
	log.Printf("[%d] @%s: %s",
		msg.Chat.ID,
		msg.From.UserName,
		msg.Text)

	// Регистрируем/обновляем пользователя
	userID, err := addOrUpdateUser(
		bot.db,
		msg.Chat.ID,
		msg.From.UserName,
		msg.From.FirstName,
		msg.From.LastName,
	)
	if err != nil {
		log.Printf("Ошибка регистрации пользователя: %v", err)
	}
	// Проверяем есть ли на данный момент этот пользователь с открытым сеансом
	bot.mesMu.Lock()
	if existing, ok := bot.activeUsers[msg.Chat.ID]; ok {
		user = existing
		fmt.Printf("Пользователь найден:\n%+v\n ", user)
	} else {
		user = &User{
			Username:  msg.From.UserName,
			FirstName: msg.From.FirstName,
			LastName:  msg.From.LastName,
			ID:        userID,
			ChatID:    msg.Chat.ID,
		}
		bot.activeUsers[msg.Chat.ID] = user
		fmt.Printf("Пользователь создан:\n%+v\n ", user)
	}
	bot.mesMu.Unlock()
	fmt.Printf("Адрес user в startRegistration: %p\nТекущая операция: %s\n", user, user.CurrentAction)

	// Сохраняем входящее сообщение
	err = addMessage(bot.db, userID, msg.Text, true, msg.IsCommand())
	if err != nil {
		log.Printf("Ошибка сохранения сообщения: %v", err)
	}

	/* Далее логика проверки и ответвления
	проверяем сообщение является командой? Если да то отправляем обрабатывать как команду
	- Проверяем есть ли активные действия по пользователю, ожидаем ли ответ.
	- обрабатываем как обычное сообщение
	*/

	// Если это команда отправляем по пути команды
	//pathCommand возвращает ответ на команду, или ошибку выполнения команды
	//отправка сообщений с клавиатурой внутри команд
	if msg.IsCommand() {
		message := pathCommand(bot, msg, user, msg.Command())
		if message != "" {
			reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Ошибка:%v", message))
			reply.ParseMode = "HTML"
			if _, err := bot.api.Send(reply); err != nil {
				log.Printf("Ошибка отправки ответа: %v", err)
				return
			}
		}
		return
	}

	//Проверяем незавершенные действия
	//скорее всего тут еще будет проверка на активный диалог с Ии, который необходимо будет завершать
	if !msg.IsCommand() { // Давай попробуем реализовать с загрузкой структуры User которая хранит текущее действие
		//Если есть открытый сеанс проверяем что делал пользователь
		if user.SessionActive {
			switch user.CurrentAction {
			case "register":
				handleRegistrationStep(bot, msg, user.ChatID)
				return
			}
		}

	}
	// Обработка простого сообщения
	fmt.Println("Обработка простого сообщения")
	response = formatDefaultText(msg)
	reply := tgbotapi.NewMessage(msg.Chat.ID, response)
	reply.ParseMode = "HTML"
	bot.mesMu.Lock()
	bot.activeUsers[msg.Chat.ID].CurrentAction = "" // сбрасываем последнее действие если простое сообщение
	bot.activeUsers[msg.Chat.ID].SessionActive = false
	bot.mesMu.Unlock()
	if _, err := bot.api.Send(reply); err != nil {
		log.Printf("Ошибка отправки ответа: %v", err)
		return
	}

	/*
		regMu.Lock()
		regData, exists := regState[msg.Chat.ID]
		regMu.Unlock()
		if exists && regData.Step != StepNone {
			handleRegistrationStep(bot, msg, regData)
			return
		}

	*/
}

func formatDefaultText(msg *tgbotapi.Message) string {
	return fmt.Sprintf(`📝 <b>Вы написали:</b> %s

Ваш ID: <code>%d</code>
Username: @%s

Используйте /help для списка команд.`,
		msg.Text,
		msg.Chat.ID,
		msg.From.UserName,
	)
}
