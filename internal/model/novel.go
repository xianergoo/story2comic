package model

import "time"

type Novel struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"index;not null"`
	Title       string    `gorm:"not null"`
	Summary     string    `gorm:"default:''"`
	CoverURL    string    `gorm:"default:''"`
	Mode        string    `gorm:"not null"`                       // inspiration / outline / blindbox
	ImageMode   string    `gorm:"not null;default:single"`        // single / multi
	Status      string    `gorm:"not null;default:drafting"`      // drafting / completed / failed / stopped
	TextStatus  string    `gorm:"not null;default:idle"`          // idle / writing / paused / done
	ImageStatus string    `gorm:"not null;default:idle"`          // idle / generating / paused / done
	AIConfigID  uint      `gorm:"default:null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
