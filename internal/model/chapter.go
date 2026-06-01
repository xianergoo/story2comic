package model

import "time"

type Chapter struct {
	ID              uint      `gorm:"primaryKey"`
	NovelID         uint      `gorm:"index;not null"`
	ChapterNo       int       `gorm:"not null"`
	Title           string    `gorm:"not null"`
	Content         string    `gorm:"default:''"`
	Status          string    `gorm:"not null;default:pending"`
	RewriteCount    int       `gorm:"default:0"`
	ContextSnapshot string    `gorm:"default:''"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
