package main

import (
	"database/sql"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

/*
Создать логику обработки сообщений
Ожидаем ли мы ответ от пользователя ChatID
если нет то обрабатываем как обычное сообщение
Иначе обрабатываем ответ и снимаем флаг ожидания ответа
*/

// Хранение состояния регистрации

type RegistrationStep int

const (
	StepCheck RegistrationStep = iota
	StepName
	StepEmail
	StepPhone
	StepComplete
	StepCorrect
)

type RegistrationData struct {
	Step   RegistrationStep
	ChatID int64
	Name   string
	Email  string
	Phone  string
}
type RegistrationManager struct {
	regState map[int64]*RegistrationData
}

type RegisterUser struct {
	RegName        string `json:"reg_name"`
	RegEmail       string `json:"reg_email"`
	RegNumberPhone string `json:"reg_number_phone"`
	//RegisterStatus RegisterStatus
	RegisterStatus map[string]bool
}

// Запуск процесса регистрации
func (bot *Bot) startRegistration(user *User) error {
	/*
		Здесь первым делом Идём в таблицу user_register проверяем регистрацию
	*/

	if _, ok := bot.regState[user.ChatID]; !ok {
		bot.regState[user.ChatID] = &RegistrationData{
			ChatID: user.ChatID,
			Step:   0,
			Name:   "",
			Email:  "",
			Phone:  "",
		}
	}
	exists, err := bot.getUserRegister(user.ChatID)

	if err != sql.ErrNoRows && err != nil {
		return err
	}
	if exists != nil {
		log.Printf("User %v already registered", user.ChatID)
		//если вытащили данные из бд, отправляем их на проверку
		if bot.regState[user.ChatID].Name != "" {
			bot.regMu.Lock()
			user.CurrentAction = "register"
			user.SessionActive = true
			registration := *bot.regState[user.ChatID]
			registration.Step = StepCheck
			bot.regMu.Unlock()
			response := fmt.Sprintf("Пользователь:%v уже зарегистрирован.\nИмя:%v \nEmail:%v \nНомер телефона:%v\nТребуется корректировка данных?", registration.ChatID, registration.Name, registration.Email, registration.Phone)
			msg := tgbotapi.NewMessage(user.ChatID, response)
			bot.api.Send(msg)
			return nil
		}
	}

	bot.regMu.Lock()
	bot.regState[user.ChatID] = &RegistrationData{
		ChatID: user.ChatID,
		Step:   StepName,
	}
	bot.regMu.Unlock()

	bot.mesMu.Lock()
	user.CurrentAction = "register"
	user.SessionActive = true
	bot.mesMu.Unlock()

	message := "Введите имя"
	reply := tgbotapi.NewMessage(user.ChatID, message)
	_, err = bot.api.Send(reply)
	if err != nil {
		return err
	}

	fmt.Println("Registration started")
	fmt.Printf("Адрес user в startRegistration: %p\n Последнеее действие: %s\n", user, user.CurrentAction)

	return nil
}

func handleRegistrationStep(bot *Bot, msg *tgbotapi.Message, UserID int64) {
	bot.regMu.Lock()
	state, exists := bot.regState[UserID]
	if !exists {
		log.Println("Не найден пользователь в статусе регистрации")
		return
	}
	step := &bot.regState[UserID].Step
	defer bot.regMu.Unlock()
	switch *step { // здесь надо мьютексы на доступ к мапе
	case StepName:
		state.Name = msg.Text
		state.Step = StepEmail
		bot.api.Send(tgbotapi.NewMessage(UserID, "Введите Email"))
	case StepEmail:
		state.Email = msg.Text
		state.Step = StepPhone
		bot.api.Send(tgbotapi.NewMessage(UserID, "Введите номер телефона"))
	case StepPhone:
		state.Phone = msg.Text
		state.Step = StepComplete
		completeRegistration(bot, state)
	case StepCheck:
		if msg.Text == "NO" {
			bot.api.Send(tgbotapi.NewMessage(UserID, "Анкета не изменена"))
			bot.clearRegistration(UserID)
		}
		if msg.Text == "YES" {
			// Очищаем мапу
			state.Step = StepName
			state.Name = ""
			state.Email = ""
			state.Phone = ""
			bot.api.Send(tgbotapi.NewMessage(UserID, "Введите имя:"))
			//bot.clearRegistration(UserID) //заглушка что бы не зависала регистрация
			// state.Step = StepCorrect
		}
		/* Две кнопки:
		OK  которая сбрасывает операцию и return,
		и Корректировка которая запускает процесс регистрации с обновлением данных в бд
		*/
	case StepCorrect:

	default:
		bot.clearRegistration(UserID)
	}
}
func completeRegistration(bot *Bot, data *RegistrationData) {
	// завершение регистрации с занесением данным в бд
	_, err := addOrUpdateUsersRegisters(bot, data)
	if err != nil {
		message := "❌ Ошибка при регистрации пользователя, повторите операцию!"
		reply := tgbotapi.NewMessage(data.ChatID, message)
		bot.clearRegistration(data.ChatID)
		if _, err := bot.api.Send(reply); err != nil {
			log.Println(err)
		}
	}
	message := fmt.Sprintf("Регистрация завершена!\nИмя:%s\nEmail:%s\nТелефон:%s\n", data.Name, data.Email, data.Phone)
	bot.api.Send(tgbotapi.NewMessage(data.ChatID, message))
	bot.clearRegistration(data.ChatID)
}

//

func (bot *Bot) clearRegistration(userID int64) {
	// здесь удалить мапу с юзером на регистрацию
	delete(bot.regState, userID)
	bot.mesMu.Lock()
	bot.activeUsers[userID].SessionActive = false
	bot.activeUsers[userID].CurrentAction = ""
	bot.mesMu.Unlock()

	fmt.Println("Registration cleared")
}

// /
func (ru RegisterUser) registerUser(msg *tgbotapi.Message) {
	/* В этой функции надо проверять был ли зарегистрирован пользователь
	если был то возвращать данные, и предложить корректировку.
	Запрос в БД по chat_id есть ли запись, если есть то возвращаем поля занося в нашу структуру и выведя её пользователю
	иначе запускаем регистрацию

	Иметь кнопку отмены которая выходит из функции без сохранения записи структуры в бд.

	*/

	/*
				В этой функции мы начинаем процесс регистрации,
			Пока что создаем структуру заполняем ее и в конце регистрации ее выводим.
		Необходимо вывести кнопки для заполнения Имя Емайл номер телефона
	*/

}

func (ru RegisterUser) statusFinish() bool {

	//for _,_ := range ru.RegisterStatus {}
	return false
}

// Функция отправки inline-клавиатуры
func (bot *Bot) sendInlineKeyboard(msg *tgbotapi.Message, str string) {
	// Создаем inline-клавиатуру
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Имя", "regName"),
			tgbotapi.NewInlineKeyboardButtonData("Email", "regEmail"),
			tgbotapi.NewInlineKeyboardButtonData("NumberPhone", "regNumberPhone")))
	newMsg := tgbotapi.NewMessage(msg.Chat.ID, str)
	newMsg.ReplyMarkup = keyboard
	_, err := bot.api.Send(newMsg)
	if err != nil {
		fmt.Printf("Ошибка отправка кнопок %v", err)
	}
}

// Обработка нажатий на inline-кнопки
func (bot *Bot) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	chatID := callbackQuery.Message.Chat.ID
	//messageID := callbackQuery.Message.MessageID
	data := callbackQuery.Data

	// Настраиваем ForceReply
	replyMarkup := tgbotapi.ForceReply{
		ForceReply: true,
		Selective:  true, // Если хотите, чтобы reply был только для конкретного пользователя
	}

	var question string

	// Ответ на callback (убираем "часики" у пользователя)
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	api, err := bot.api.Request(callback)
	fmt.Printf("Callback ______ %v\n", api)
	if err != nil {
		fmt.Printf("ошибка обработки calback %v", err)
	}

	switch data {
	case "regName":
		question = "Пожалуйста, введите ваше имя:"

	case "regEmail":
		question = "Пожалуйста, введите ваш email:"

	case "regNumberPhone":
		question = "Пожалуйста, введите ваш номер телефона:"

	}
	msg := tgbotapi.NewMessage(chatID, question)
	msg.ReplyMarkup = replyMarkup
	bot.api.Send(msg)
}
