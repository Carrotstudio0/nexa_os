package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Store is a thread-safe key-value store with persistence
type Store struct {
	mu       sync.RWMutex
	Data     map[string]string `json:"data"`
	filename string
}

// NewStore creates or loads a store from disk
func NewStore(filename string) (*Store, error) {
	s := &Store{
		Data:     make(map[string]string),
		filename: filename,
	}

	// Try to load existing data
	if _, err := os.Stat(filename); err == nil {
		content, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		if len(content) > 0 {
			if err := json.Unmarshal(content, &s.Data); err != nil {
				return nil, fmt.Errorf("failed to parse storage file: %v", err)
			}
		}
	}

	return s, nil
}

// Set saves a value and persists to disk
func (s *Store) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data[key] = value
	return s.save()
}

// Get retrieves a value
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.Data[key]
	return val, ok
}

// List returns all keys
func (s *Store) List() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.Data))
	for k := range s.Data {
		keys = append(keys, k)
	}
	return keys
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filename, data, 0644)
}
