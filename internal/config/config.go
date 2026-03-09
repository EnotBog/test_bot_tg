package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	TelegramBotToken string  `yaml:"telegram_bot_token"`
	DatabaseURL      string  `yaml:"database_url"`
	DebugMode        bool    `yaml:"debug_mode"`
	AdminIDs         []int64 `yaml:"admin_ids"`
}

// Загружает конфигурацию из файла или переменных окружения
func LoadConfig(filename string) (Config, error) {
	var config Config
	if fileExist(filename) {
		data, err := os.ReadFile(filename)
		if err != nil {
			return config, fmt.Errorf("ошибка чтения файла конфигурации %v", err)
		}
		if ext := filepath.Ext(filename); ext == ".yaml" || ext == ".yml" {
			log.Println("Загрузка конфигурации yaml")
			err = yaml.Unmarshal(data, &config)
		} else {
			err = json.Unmarshal(data, &config)
			if err != nil {
				return config, fmt.Errorf("ошибка парсинга конфигурации %v", err)
			}
		}
		log.Printf("Конфигурация загружена из %s\n", filename)
		// Получаем токен бота пробуем из файла ежи нет, тогда из окружения
		err = godotenv.Load()
		if err != nil {
			log.Printf("Ошибка загрузки окружения из.env file: %v", err)
		}
		if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
			config.TelegramBotToken = token
		} else {
			log.Fatal("Ошибка: TELEGRAM_BOT_TOKEN не установлен")
		}
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
