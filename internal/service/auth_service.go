package service

import (
	"errors"
	"novelforge/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct{ db *gorm.DB }

func NewAuthService(db *gorm.DB) *AuthService { return &AuthService{db} }

func (s *AuthService) Register(username, password string) (*model.User, error) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := model.User{Username: username, PasswordHash: string(hash)}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, errors.New("用户名已存在")
	}
	return &user, nil
}

func (s *AuthService) Login(username, password string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	return &user, nil
}
