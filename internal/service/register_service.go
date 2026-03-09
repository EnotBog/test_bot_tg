package service

import (
	"telegram-bot/internal/models"
	"telegram-bot/internal/repository"
)

type RegisterService struct {
	regRepo *repository.RegisterRepository
}

func NewRegisterService(regRepo *repository.RegisterRepository) *RegisterService {
	return &RegisterService{regRepo: regRepo}
}

func (s *RegisterService) Save(data *models.RegistrationData) error {
	return s.regRepo.CreateOrUpdate(data)
}

func (s *RegisterService) GetByChatID(chatID int64) (*models.RegistrationData, error) {
	return s.regRepo.GetByChatID(chatID)
}
