package model

import "time"

type ComicPage struct {
	ID              uint      `gorm:"primaryKey"`
	ChapterID       uint      `gorm:"index;not null"`
	NovelID         uint      `gorm:"index;not null"`
	PageNo          int       `gorm:"not null"`
	PanelCount      int       `gorm:"not null;default:4"`
	Script          string    `gorm:"default:'{}'"`
	ImageURLs       string    `gorm:"default:'[]'"`
	StyleDesc       string    `gorm:"default:''"`
	Status          string    `gorm:"not null;default:pending"`
	RetryCount      int       `gorm:"default:0"`
	ContextSnapshot string    `gorm:"default:''"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
