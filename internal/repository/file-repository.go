package repository

import (
	"encoding/json"
	"fmt"
	"os"
)

type FileRepositoryURL struct {
	filepath string
}

func NewFileRepositoryURL(filepath string) *FileRepositoryURL {
	return &FileRepositoryURL{
		filepath: filepath,
	}
}

func (r *FileRepositoryURL) Add(url string, id string) error {
	file, err := os.OpenFile(r.filepath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	entry := map[string]string{id: url}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte(data))
	if err != nil {
		return err
	}
	_, err = file.Write([]byte("\n"))
	if err != nil {
		return err
	}

	return nil
}

func (r *FileRepositoryURL) GetByID(id string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (r *FileRepositoryURL) Delete(id string) error {
	return fmt.Errorf("вызван не поддерживаемый метод FileRepository.Delete")
}
