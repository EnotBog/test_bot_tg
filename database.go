package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

// Инициализация базы данных
func initDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./bot_database.db")
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %v", err)
	}

	createUserTable := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER UNIQUE NOT NULL,
		username TEXT,
		first_name TEXT,
		last_name TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_active DATETIME DEFAULT CURRENT_TIMESTAMP,
		is_active BOOLEAN DEFAULT TRUE
	);`

	createUserRegisterTable := `CREATE TABLE IF NOT EXISTS users_registers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id INTEGER UNIQUE NOT NULL,
    name TEXT,
    email TEXT,
    number_phone INTEGER UNIQUE NOT NULL,
    FOREIGN KEY (chat_id) REFERENCES users(chat_id) ON DELETE CASCADE
);`

	// Создаем таблицу сообщений
	createMessagesTable := `
    CREATE TABLE IF NOT EXISTS messages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        message_text TEXT NOT NULL,
        is_from_user BOOLEAN DEFAULT TRUE,
        is_command BOOLEAN DEFAULT FALSE,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );`

	// Создаем таблицы
	for _, query := range []string{createUserTable, createMessagesTable, createUserRegisterTable} {
		_, err := db.Exec(query)
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("Ошибка создания таблицы: %v\n %s \n", err, query)
		}
	}

	// Включаем поддержку внешних ключей
	db.Exec("PRAGMA foreign_keys = ON")

	return db, nil
}

// Добавление или обновление пользователя
func addOrUpdateUser(db *sql.DB, chatID int64, username, firstName, lastName string) (int64, error) {
	var userID int64

	// Проверяем, существует ли пользователь
	err := db.QueryRow(`SELECT id FROM users WHERE chat_id = ?`, chatID).Scan(&userID)
	if err == nil {
		// Пользователь существует, обновляем информацию
		_, err = db.Exec(`UPDATE users SET username = ?, first_name = ?, last_name = ?,last_active = ? WHERE chat_id = ?`,
			username, firstName, lastName, time.Now(), chatID)
		if err != nil {
			return 0, fmt.Errorf("ошибка обновления пользователя: %v", err)
		}
		log.Printf("Пользователь обновлен: ChatID=%d, ID=%d", chatID, userID)
		return userID, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("ошибка проверки пользователя: %v", err)
	}

	// Пользователь не существует, создаем нового

	resut, err := db.Exec(`
        INSERT INTO users (chat_id, username, first_name, last_name, created_at, last_active)
        VALUES (?, ?, ?, ?, ?, ?)`, chatID, username, firstName, lastName, time.Now(), time.Now())
	if err != nil {
		return 0, fmt.Errorf("ошибка создания пользователя: %v, %v", username, err)
	}

	userID, err = resut.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения userID,%v", err)
	}

	log.Printf("Новый пользователь: ChatID=%d, ID=%d", chatID, userID)
	return userID, nil
}

// Добавление Обновление пользователя в БД
func addMessage(db *sql.DB, userID int64, text string, isFromUser, isCommand bool) error {
	_, err := db.Exec(
		`INSERT INTO messages (user_id,message_text,is_from_user,is_command,created_at)
				VAlUES(?,?,?,?,?)`, userID, text, isFromUser, isCommand, time.Now())
	if err != nil {
		return fmt.Errorf("ошибка сохранения сообщения: %v", err)
	}
	return nil
}

// Получение сообщений из таблицы messages
func getUserMessages(db *sql.DB, userID int64, limit int) ([]string, error) {
	var messages []string

	rows, err := db.Query(`
        SELECT message_text, is_from_user, created_at
        FROM messages
        WHERE user_id = ?
        ORDER BY created_at DESC
        LIMIT ?`,
		userID, limit)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сообщений: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var text string
		var isFromUser bool
		var createdAt time.Time
		err := rows.Scan(&text, &isFromUser, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения сообщения: %v", err)
		}

		prefix := "👤"
		if !isFromUser {
			prefix = "🤖"
		}
		messages = append(messages, fmt.Sprintf("%s %s: %s",
			prefix, createdAt.Format("15:04"), text))
	}

	return messages, nil
}

// Получение статистики
func getStats(db *sql.DB) (int, int, error) {
	var userCount, messageCount int

	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		return 0, 0, fmt.Errorf("ошибка подсчета пользователей: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&messageCount)
	if err != nil {
		return 0, 0, fmt.Errorf("ошибка подсчета сообщений: %v", err)
	}

	return userCount, messageCount, nil
}
