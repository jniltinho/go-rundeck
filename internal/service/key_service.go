package service

import (
	"go-rundeck/internal/model"
	"go-rundeck/internal/repository"
)

type KeyService struct {
	repo *repository.KeyRepository
}

func NewKeyService(repo *repository.KeyRepository) *KeyService {
	return &KeyService{repo: repo}
}

func (s *KeyService) Create(name string, keyType model.KeyType, content string, description string, projectID *uint, userID uint) (*model.KeyStorage, error) {
	encrypted, err := s.repo.Encrypt(content)
	if err != nil {
		return nil, err
	}

	key := &model.KeyStorage{
		Name:        name,
		KeyType:     keyType,
		ContentEnc:  encrypted,
		Description: description,
		ProjectID:   projectID,
		CreatedBy:   userID,
	}

	if err := s.repo.Create(key); err != nil {
		return nil, err
	}
	return key, nil
}

func (s *KeyService) ListSystemKeys() ([]model.KeyStorage, error) {
	return s.repo.ListByProject(nil)
}

func (s *KeyService) GetDecryptedContent(id uint) (string, error) {
	key, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}
	return s.repo.Decrypt(key.ContentEnc)
}

func (s *KeyService) Update(id uint, name string, keyType model.KeyType, description string, newContent string) error {
	key, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if name != "" {
		key.Name = name
	}
	key.KeyType = keyType
	key.Description = description
	if newContent != "" {
		encrypted, err := s.repo.Encrypt(newContent)
		if err != nil {
			return err
		}
		key.ContentEnc = encrypted
	}
	return s.repo.Update(key)
}

func (s *KeyService) Delete(id uint) error {
	return s.repo.Delete(id)
}
