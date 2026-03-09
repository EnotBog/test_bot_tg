package main

import (
	"fmt"
	"log"

	"telegram-bot/internal/bot"
	"telegram-bot/internal/config"
	"telegram-bot/internal/database"
	"telegram-bot/internal/handler"
	"telegram-bot/internal/repository"
	"telegram-bot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	fmt.Println("=== Запуск Telegram бота ===")

	// Загрузка конфигурации
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("error load config. %v\n", err)
	}

	// Подключение к БД
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("error connect to database. %v\n", err)
	}
	defer db.Close()

	// Инициализация базы данных
	if err := database.Init(db); err != nil {
		log.Fatalf("error init database. %v\n", err)
	}

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db)
	msgRepo := repository.NewMessageRepository(db)
	regRepo := repository.NewRegisterRepository(db)

	// Инициализация сервисов
	userService := service.NewUserService(userRepo)
	msgService := service.NewMessageService(msgRepo)
	regService := service.NewRegisterService(regRepo)

	// Инициализация хендлера
	h := handler.NewHandler(userService, msgService, regService)

	// Создание Telegram бота
	api, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	api.Debug = cfg.DebugMode

	// Создание бота
	tgBot := bot.NewBot(api, &cfg, h, userService, msgService)
	//tgBot := bot.New(api, cfg, h, userService, msgService)

	// Запуск
	fmt.Printf("Бот @%s запущен\n", api.Self.UserName)
	tgBot.Start()
}
