package service

import (
	"fmt"
	"telegram-bot/internal/models"
	"telegram-bot/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetOrCreate(chatID int64, username, firstName, lastName string) (*models.User, error) {
	id, err := s.userRepo.CreateOrUpdate(chatID, username, firstName, lastName)
	if err != nil {
		return nil, fmt.Errorf("userRepo.CreateOrUpdate: %w\n", err)
	}

	user, err := s.userRepo.GetByChatID(chatID)
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetByChatID: %w\n", err)
	}
	user.ID = id
	return user, nil
}

func (s *UserService) GetStats() (int, int, error) {
	return s.userRepo.GetStats()
}
