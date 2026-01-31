package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTrustedUsersStore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filePath string
	}{
		{"simple path", "trusted_users.json"},
		{"path with directory", "/tmp/trusted_users.json"},
		{"relative path", "./data/trusted_users.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := NewTrustedUsersStore(tt.filePath)
			require.NotNil(t, store, "NewTrustedUsersStore() returned nil")
			assert.Equal(t, filepath.Clean(tt.filePath), store.filePath, "filePath mismatch")
		})
	}
}

func TestTrustedUsersStore_LoadEmpty(t *testing.T) {
	t.Parallel()

	store := NewTrustedUsersStore("non_existent_file.json")
	users, err := store.Load()

	require.NoError(t, err, "Load() should not error for non-existent file")
	assert.Empty(t, users, "Load() should return empty map for non-existent file")
}

func TestTrustedUsersStore_SaveAndLoad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		users map[string]bool
	}{
		{
			name:  "empty map",
			users: map[string]bool{},
		},
		{
			name: "single user",
			users: map[string]bool{
				"testuser": true,
			},
		},
		{
			name: "multiple users",
			users: map[string]bool{
				"user1":    true,
				"user2":    true,
				"user3":    true,
				"frankos6": true,
			},
		},
		{
			name: "users with special chars",
			users: map[string]bool{
				"user_with_underscore": true,
				"user123":              true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "trusted_users.json")
			store := NewTrustedUsersStore(filePath)

			err := store.Save(tt.users)
			require.NoError(t, err, "Save() error")

			_, err = os.Stat(filePath)
			require.NoError(t, err, "Save() did not create file")

			loaded, err := store.Load()
			require.NoError(t, err, "Load() error")

			assert.Len(t, loaded, len(tt.users), "Load() returned wrong number of users")

			for user := range tt.users {
				assert.True(t, loaded[user], "Load() missing user %q", user)
			}
		})
	}
}

func TestTrustedUsersStore_LoadInvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "trusted_users.json")

	err := os.WriteFile(filePath, []byte("not valid json"), 0644)
	require.NoError(t, err, "Could not write test file")

	store := NewTrustedUsersStore(filePath)
	_, err = store.Load()

	assert.Error(t, err, "Load() should return error for invalid JSON")
}

func TestTrustedUsersStore_SaveOverwrite(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "trusted_users.json")
	store := NewTrustedUsersStore(filePath)

	initial := map[string]bool{"user1": true, "user2": true}
	err := store.Save(initial)
	require.NoError(t, err, "Save() initial error")

	updated := map[string]bool{"user3": true}
	err = store.Save(updated)
	require.NoError(t, err, "Save() updated error")

	loaded, err := store.Load()
	require.NoError(t, err, "Load() error")

	assert.Len(t, loaded, 1, "Load() returned wrong number of users")
	assert.True(t, loaded["user3"], "Load() missing user3")
	assert.False(t, loaded["user1"], "Load() should not contain user1 after overwrite")
}

func TestResolveTrustedUsersPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		envPath string
		want    string
	}{
		{
			name:    "simple path",
			envPath: ".env",
			want:    "trusted_users.json",
		},
		{
			name:    "path with directory",
			envPath: filepath.FromSlash("/home/user/.env"),
			want:    filepath.FromSlash("/home/user/trusted_users.json"),
		},
		{
			name:    "windows style path",
			envPath: filepath.FromSlash("C:/Users/test/.env"),
			want:    filepath.FromSlash("C:/Users/test/trusted_users.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, ResolveTrustedUsersPath(tt.envPath), "ResolveTrustedUsersPath(%q)", tt.envPath)
		})
	}
}
