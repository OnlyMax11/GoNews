package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"GoNews/pkg/api"
	"GoNews/pkg/config"
	"GoNews/pkg/db"
	"GoNews/pkg/model"
	"GoNews/pkg/rss"
)

func main() {
	// Инициализация логгера
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting application...")

	// Загрузка конфигурации
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключение к PostgreSQL
	dsn := "user=usergonews password=rcs58w58 dbname=gonews host=localhost sslmode=disable"
	storage, err := db.New(dsn)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()
	log.Println("Successfully connected to database")

	// Создание API сервера
	apiServer := api.New(storage)

	// Настройка HTTP сервера
	srv := &http.Server{
		Addr:    ":8080",
		Handler: apiServer.Router(),
	}

	// Контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запуск фонового агрегатора новостей
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runAggregator(ctx, cfg, storage)
	}()

	// Запуск HTTP сервера в отдельной горутине
	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Ожидание сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	cancel() // Остановка агрегатора
	wg.Wait()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Application stopped")
}

func runAggregator(ctx context.Context, cfg *config.Config, storage *db.Storage) {
	log.Println("Starting news aggregator...")
	ticker := time.NewTicker(time.Duration(cfg.RequestPeriod) * time.Minute)
	defer ticker.Stop()

	// Первый запуск сразу после старта
	processFeeds(cfg.RSS, storage)

	for {
		select {
		case <-ticker.C:
			processFeeds(cfg.RSS, storage)
		case <-ctx.Done():
			log.Println("Stopping news aggregator...")
			return
		}
	}
}

func processFeeds(feeds []string, storage *db.Storage) {
	log.Printf("Processing %d feeds", len(feeds))
	var wg sync.WaitGroup
	postsCh := make(chan model.Post, 100) // Буферизованный канал
	errCh := make(chan error, len(feeds))

	for _, url := range feeds {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			posts, err := rss.Parse(url)
			if err != nil {
				errCh <- fmt.Errorf("error parsing %s: %w", url, err)
				return
			}
			for _, post := range posts {
				postsCh <- post
			}
		}(url)
	}

	// Закрытие каналов после завершения всех горутин
	go func() {
		wg.Wait()
		close(postsCh)
		close(errCh)
	}()

	// Обработка результатов
	savedCount := 0
	for post := range postsCh {
		if err := storage.SavePost(post); err != nil {
			log.Printf("Error saving post: %v", err)
			continue
		}
		savedCount++
	}

	for err := range errCh {
		log.Printf("RSS error: %v", err)
	}

	log.Printf("Saved %d new posts", savedCount)
}
