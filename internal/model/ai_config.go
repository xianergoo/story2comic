package model

import "time"

type AIConfig struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"index;not null"`
	Name       string    `gorm:"not null"`
	Provider   string    `gorm:"not null"` // openai / qwen / custom
	APIKey     string    `gorm:"not null"`
	BaseURL    string    `gorm:"default:''"`
	TextModel  string    `gorm:"not null"`
	ImageModel string    `gorm:"not null"`
	IsDefault  bool      `gorm:"default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
