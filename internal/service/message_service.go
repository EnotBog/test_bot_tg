package service

import (
	"fmt"
	"telegram-bot/internal/models"
	"telegram-bot/internal/repository"
)

type MessageService struct {
	msgRepo *repository.MessageRepository
}

func NewMessageService(msgRepo *repository.MessageRepository) *MessageService {
	return &MessageService{msgRepo: msgRepo}
}

func (s *MessageService) Save(chatID int64, text string, isFromUser, isCommand bool) error {
	err := s.msgRepo.Save(chatID, text, isFromUser, isCommand)
	if err != nil {
		return fmt.Errorf("save message: %w\n", err)
	}
	return nil
}

func (s *MessageService) GetHistory(chatID int64, limit int) ([]models.Message, error) {
	return s.msgRepo.GetByChatID(chatID, limit)
}
