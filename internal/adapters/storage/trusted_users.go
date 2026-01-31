package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type TrustedUsersStore struct {
	filePath string
	mu       sync.Mutex
}

type TrustedUsersData struct {
	Users []string `json:"users"`
}

func NewTrustedUsersStore(filePath string) *TrustedUsersStore {
	return &TrustedUsersStore{
		filePath: filepath.Clean(filePath),
	}
}

func (s *TrustedUsersStore) Load() (map[string]bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	users := make(map[string]bool)

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return empty map (not an error)
			return users, nil
		}
		return nil, err
	}

	var stored TrustedUsersData
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, err
	}

	for _, user := range stored.Users {
		users[user] = true
	}

	return users, nil
}

func (s *TrustedUsersStore) Save(users map[string]bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userList := make([]string, 0, len(users))
	for user := range users {
		userList = append(userList, user)
	}

	data := TrustedUsersData{Users: userList}
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, "trusted_users.tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(jsonData); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return err
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, s.filePath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return nil
}

func ResolveTrustedUsersPath(envPath string) string {
	dir := filepath.Dir(envPath)
	return filepath.Join(dir, "trusted_users.json")
}
