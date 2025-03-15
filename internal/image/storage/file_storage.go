package storage

import (
	"fmt"
	"os"
	"path/filepath"
    "context"

	"github.com/Nzyazin/zadnik.store/internal/image/domain"
)

type fileStorage struct {
    basePath string
    baseURL string
}

func NewFileStorage(basePath string, baseURL string) (domain.ImageStorage, error) {
    if err := os.MkdirAll(basePath, 0755); err != nil {
        return nil, fmt.Errorf("failed to create storage directory: %w", err)
    }

    return &fileStorage{
        basePath: basePath,
        baseURL: baseURL,
    }, nil
}

func (fs *fileStorage) Store(ctx context.Context, imageData []byte, productID int32) (string, error) {
    filename := fmt.Sprintf("%d.jpg", productID)
    filePath := filepath.Join(fs.basePath, filename)
    
    if err := os.WriteFile(filePath, imageData, 0644); err != nil {
        return "", fmt.Errorf("failed to write image file: %w", err)
    }

    return fmt.Sprintf("%s/%s", fs.baseURL, filename), nil
}

func (fs *fileStorage) Delete(ctx context.Context, imageURL string) error {
    filename := filepath.Base(imageURL)
    filePath := filepath.Join(fs.basePath, filename)

    if err := os.Remove(filePath); err != nil {
        if os.IsNotExist(err) {
            return nil
        }
        return fmt.Errorf("failed to delete image file: %w", err)
    }

    return nil
}

func (fs *fileStorage) GetBaseURL() string {
    return fs.baseURL
}
