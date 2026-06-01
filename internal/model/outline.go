package model

import "time"

type Outline struct {
	ID              uint      `gorm:"primaryKey"`
	NovelID         uint      `gorm:"index;not null"`
	Version         int       `gorm:"not null;default:1"`
	Content         string    `gorm:"not null"`
	CharacterSheets string    `gorm:"default:'{}'"`
	WorldSetting    string    `gorm:"default:'{}'"`
	ChapterPlan     string    `gorm:"default:'[]'"`
	CreatedAt       time.Time
}
