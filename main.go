package main

import (
	"database/sql"
	"fmt"
	_ "github.com/joho/godotenv"
	"log"
	_ "os"
	_ "strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	regState    map[int64]*RegistrationData
	activeUsers map[int64]*User
	regMu       sync.RWMutex
	mesMu       sync.RWMutex
	api         *tgbotapi.BotAPI
	db          *sql.DB
}

func main() {
	fmt.Println("=== Запуск Telegram бота с базой данных ===")

	// Инициализация базы данных
	var err error

	dbCreate, err := initDatabase()
	if err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(dbCreate)

	// Загрузка конфигурации
	config, err := loadConfig("config.xml")
	if err != nil {
		log.Fatalf("ошибка конфигурации %v", err)
	}

	// Создаем бота
	botCreate, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	botCreate.Debug = config.DebugMode // Отключаем отладку в проде

	// Инициализируем структуру бота
	botAI := &Bot{
		api:         botCreate,
		db:          dbCreate,
		regState:    make(map[int64]*RegistrationData),
		activeUsers: make(map[int64]*User),
	}

	// Получаем информацию о боте
	botInfo, _ := botAI.api.GetMe()
	fmt.Printf("Бот запущен: @%s\n", botInfo.UserName)

	// Выводим статистику при запуске
	userCount, messageCount, err := getStats(botAI.db)
	if err != nil {
		log.Printf("Ошибка получения статистики: %v", err)
	} else {
		fmt.Printf("Статистика: %d пользователей, %d сообщений\n", userCount, messageCount)
	}

	// Настраиваем поллинг
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := botAI.api.GetUpdatesChan(updateConfig)

	// Обработка сообщений
	for update := range updates {

		if update.CallbackQuery != nil {
			go botAI.handleCallbackQuery(update.CallbackQuery)
		}
		if update.Message != nil && update.Message.Text != "" {
			go botAI.handleIncomingMessage(update)
		}
	}
}
