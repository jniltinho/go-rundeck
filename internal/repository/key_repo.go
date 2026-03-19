package repository

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"go-rundeck/internal/model"

	"gorm.io/gorm"
)

type KeyRepository struct {
	db     *gorm.DB
	secret []byte // Needs to be 32 bytes for AES-256
}

func NewKeyRepository(db *gorm.DB, secret string) (*KeyRepository, error) {
	if len(secret) < 32 {
		return nil, errors.New("secret key must be at least 32 bytes")
	}
	// Use exactly 32 bytes for AES-256
	keyBuf := make([]byte, 32)
	copy(keyBuf, []byte(secret))
	return &KeyRepository{db: db, secret: keyBuf}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (r *KeyRepository) Encrypt(plaintext string) (string, error) {
	c, err := aes.NewCipher(r.secret)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt decrypts AES-256-GCM cipher text
func (r *KeyRepository) Decrypt(cipherTextB64 string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(cipherTextB64)
	if err != nil {
		return "", err
	}

	c, err := aes.NewCipher(r.secret)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func (r *KeyRepository) ListByProject(projectID *uint) ([]model.KeyStorage, error) {
	var keys []model.KeyStorage
	query := r.db
	if projectID != nil {
		query = query.Where("project_id = ?", projectID)
	} else {
		query = query.Where("project_id IS NULL")
	}
	err := query.Order("created_at asc").Find(&keys).Error
	return keys, err
}

func (r *KeyRepository) GetByID(id uint) (*model.KeyStorage, error) {
	var key model.KeyStorage
	if err := r.db.First(&key, id).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *KeyRepository) Create(key *model.KeyStorage) error {
	return r.db.Create(key).Error
}

func (r *KeyRepository) Delete(id uint) error {
	// Unlink any nodes referencing this key before deleting
	if err := r.db.Exec("UPDATE nodes SET key_id = NULL WHERE key_id = ?", id).Error; err != nil {
		return err
	}
	return r.db.Delete(&model.KeyStorage{}, id).Error
}
