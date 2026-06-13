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

// func TestURLService_GetOriginalURLByID(t *testing.T) {
// 	cfg := config.Config{
// 		ServerAddress:       ":8080",
// 		BaseShortURLAddress: "http://localhost:8080",
// 	}
// 	tests := []struct {
// 		name    string
// 		cfg     config.Config
// 		id      string
// 		want    model.ShortenModel
// 		wantErr error
// 	}{
// 		{
// 			name:    "check that the value received from the repository is being returned",
// 			cfg:     cfg,
// 			id:      "DFSDFDD",
// 			want:    model.ShortenModel{ID: "test", OriginURL: "http:/test.ru", UserID: "test", DeletedFlag: false},
// 			wantErr: nil,
// 		},
// 		{
// 			name:    "check that an error is returned if an error has occurred in the repository",
// 			cfg:     cfg,
// 			id:      "DFSDFDD",
// 			want:    model.ShortenModel{ID: "test", OriginURL: "http:/test.ru", UserID: "test", DeletedFlag: false},
// 			wantErr: errors.New("test"),
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			m := mocks.NewMockURLRepository(ctrl)

// 			m.EXPECT().GetByID(context.Background(), tt.id).Times(1).Return(tt.want, tt.wantErr)

// 			g := service.URLGenerator{}
// 			us := service.NewURLService(m, tt.cfg, g)

// 			got, gotErr := us.GetOriginalURLByID(context.Background(), tt.id)

// 			if tt.wantErr != nil {
// 				assert.NotNil(t, gotErr)
// 			} else {
// 				assert.Equal(t, got, tt.want)
// 			}

// 		})
// 	}
// }

// func TestURLService_MakeShortURLByID(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		cfg     config.Config
// 		id      string
// 		want    string
// 		wantErr error
// 	}{
// 		{
// 			name: "Check the correctness of the operation when the base URL ends /",
// 			cfg: config.Config{
// 				ServerAddress:       ":8080",
// 				BaseShortURLAddress: "http://test.ru/",
// 			},
// 			id:      "DFSDFDD",
// 			want:    "http://test.ru/DFSDFDD",
// 			wantErr: nil,
// 		},
// 		{
// 			name: "Check the correctness of the operation when the base URL ends without /",
// 			cfg: config.Config{
// 				ServerAddress:       ":8080",
// 				BaseShortURLAddress: "http://test.ru",
// 			},
// 			id:      "DFSDFDD",
// 			want:    "http://test.ru/DFSDFDD",
// 			wantErr: nil,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			m := mocks.NewMockURLRepository(ctrl)
// 			g := service.URLGenerator{}
// 			us := service.NewURLService(m, tt.cfg, g)

// 			got, gotErr := us.MakeShortURLByID(context.Background(), tt.id)

// 			if tt.wantErr != nil {
// 				assert.NotNil(t, gotErr)
// 			} else {
// 				assert.Equal(t, got, tt.want)
// 			}

//			})
//		}
//	}
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
