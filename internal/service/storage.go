package service

import (
	"encoding/json"
	"os"
	"scrapify/pkg/models"
	"sync"
)

type StorageService struct {
	filePath string
	mu       sync.RWMutex
}

func NewStorageService(filePath string) *StorageService {
	// Initialize file if not exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.WriteFile(filePath, []byte("[]"), 0644)
	}
	return &StorageService{filePath: filePath}
}

func (s *StorageService) SaveTask(task *models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.readTasks()
	if err != nil {
		return err
	}

	exists := false
	for i, t := range tasks {
		if t.ID == task.ID {
			tasks[i] = task
			exists = true
			break
		}
	}

	if !exists {
		tasks = append(tasks, task)
	}

	return s.writeTasks(tasks)
}

func (s *StorageService) GetTask(id string) (*models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks, err := s.readTasks()
	if err != nil {
		return nil, err
	}

	for _, t := range tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, nil
}

func (s *StorageService) readTasks() ([]*models.Task, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}
	var tasks []*models.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *StorageService) writeTasks(tasks []*models.Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}
