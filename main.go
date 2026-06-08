package main

import (
	"context"
	"log"
	"novelforge/internal/app"
	"novelforge/internal/config"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()
	application, err := app.New(cfg)
	if err != nil {
		log.Fatal("启动失败:", err)
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-sigCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := application.Shutdown(shutdownCtx); err != nil {
			log.Printf("关闭失败: %v", err)
		}
	}()

	if err := application.Run(); err != nil {
		log.Fatal("运行失败:", err)
	}
}
