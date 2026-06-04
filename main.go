package main

import (
	"log"
	"novelforge/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatal("启动失败:", err)
	}
	if err := application.Run(); err != nil {
		log.Fatal("运行失败:", err)
	}
}
