package main

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"novelforge/internal/config"
	"novelforge/internal/model"
)

func initDB(dbPath string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	db.AutoMigrate(
		&model.User{},
		&model.AIConfig{},
		&model.Novel{},
		&model.Outline{},
		&model.Chapter{},
		&model.ComicPage{},
	)
	return db
}

func main() {
	cfg := config.Load()
	db := initDB(cfg.DBPath)
	fmt.Printf("DB migrated successfully, port=%s db=%s\n", cfg.Port, cfg.DBPath)
	_ = db
}
