package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Config struct {
	TelegramBotToken string  `json:"telegram_bot_token"`
	DebugMode        bool    `json:"debug_mode"`
	AdminIDs         []int64 `json:"admin_ids"`
}

// Загружает конфигурацию из файла или переменных окружения
func loadConfig(filename string) (Config, error) {
	var config Config
	if fileExist(filename) {
		data, err := os.ReadFile(filename)
		if err != nil {
			return config, fmt.Errorf("ошибка чтения файла конфигурации %v", err)
		}
		err = json.Unmarshal(data, &config)
		if err != nil {
			return config, fmt.Errorf("ошибка парсинга конфигурации %v", err)
		}
		log.Printf("Конфигурация загружена из %s\n", filename)
		return config, nil
	}
	log.Printf("Загрузка конфигурации из файла не удалась, загрузка дефолтных значений/\n")

	// Получаем токен бота пробуем из файла ежи нет, тогда из окружения
	err := godotenv.Load()
	if err != nil {
		log.Printf("Ошибка загрузки окружения из.env file: %v", err)
	}
	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		config.TelegramBotToken = token
	} else {
		log.Fatal("Ошибка: TELEGRAM_BOT_TOKEN не установлен")
	}
	if admin := os.Getenv("ADMIN_BOT"); admin != "" {
		adminID, err := strconv.ParseInt(admin, 10, 64)
		if err != nil {
			log.Printf("Ошибка преобразования в число:%s Ошибка:%v", admin, err)
		}
		config.AdminIDs = append(config.AdminIDs, adminID)
	} else {
		log.Printf("Ошибка: ADMIN_BOT не установлен")
	}
	config.DebugMode = false

	fmt.Printf("----Создана конфигурация----\n"+
		" BOT_Token: %v\n Debug: %v\n Администраторы: %v\n"+
		"-----------------------------\n",
		"******", config.DebugMode, config.AdminIDs)
	return config, nil
}

func fileExist(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func creatDefaultConfig() {

}
