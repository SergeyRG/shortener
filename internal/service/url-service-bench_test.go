package service_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/SergeyRG/shortener/internal/config"
	"github.com/SergeyRG/shortener/internal/model"
	"github.com/SergeyRG/shortener/internal/repository"
	"github.com/SergeyRG/shortener/internal/service"
)

func fakeURL() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("https://%s.test", hex.EncodeToString(b))
}

func generateFakeURLs(count int64) []string {
	urls := make([]string, count)
	for i := range count {
		urls[i] = fakeURL()
	}
	return urls
}

func BenchmarkService_AddShortURL(b *testing.B) {
	count := 1000000
	urls := generateFakeURLs(int64(count))
	cfg := config.Config{
		ServerAddress:       ":8080",
		BaseShortURLAddress: "http://localhost:8080",
		FileStoragePath:     filepath.Join(b.TempDir(), "test_storage.txt"),
	}

	if err := os.WriteFile(cfg.FileStoragePath, []byte(""), 0644); err != nil {
		b.Fatalf("не удалось подготовить файл хранилища: %v", err)
	}

	var repo service.URLRepository
	stor := make(map[string]*model.ShortenModel)
	repo, err := repository.NewInMemoryRepositoryURL(stor, cfg.FileStoragePath)
	if err != nil {
		b.Fatalf("не удалось создать объект репозиториия: %v", err)
	}

	g := service.URLGenerator{}
	svc := service.NewURLService(repo, cfg, g)
	j := 0
	runtime.GC()
	b.ResetTimer()
	for i := 0; i <= b.N; i++ {
		svc.AddShortURL(b.Context(), urls[j], "test")
		j = (j + 1) % count
	}
	repo.Close()
}
