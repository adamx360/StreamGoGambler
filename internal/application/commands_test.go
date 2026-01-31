package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fullMsg  string
		prefix   string
		wantCmd  string
		wantArgs []string
		wantOK   bool
	}{
		{
			name:     "simple command no args",
			fullMsg:  "!status",
			prefix:   "!",
			wantCmd:  "status",
			wantArgs: nil,
			wantOK:   true,
		},
		{
			name:     "command with one arg",
			fullMsg:  "!heist 1000",
			prefix:   "!",
			wantCmd:  "heist",
			wantArgs: []string{"1000"},
			wantOK:   true,
		},
		{
			name:     "command with multiple args",
			fullMsg:  "!trust user1 user2",
			prefix:   "!",
			wantCmd:  "trust",
			wantArgs: []string{"user1", "user2"},
			wantOK:   true,
		},
		{
			name:     "command is lowercased",
			fullMsg:  "!STATUS",
			prefix:   "!",
			wantCmd:  "status",
			wantArgs: nil,
			wantOK:   true,
		},
		{
			name:     "args preserve case",
			fullMsg:  "!trust UserName",
			prefix:   "!",
			wantCmd:  "trust",
			wantArgs: []string{"UserName"},
			wantOK:   true,
		},
		{
			name:     "different prefix",
			fullMsg:  "?help",
			prefix:   "?",
			wantCmd:  "help",
			wantArgs: nil,
			wantOK:   true,
		},
		{
			name:     "multi-char prefix",
			fullMsg:  "bot!status",
			prefix:   "bot!",
			wantCmd:  "status",
			wantArgs: nil,
			wantOK:   true,
		},
		{
			name:     "wrong prefix returns false",
			fullMsg:  "!status",
			prefix:   "?",
			wantCmd:  "",
			wantArgs: nil,
			wantOK:   false,
		},
		{
			name:     "empty message returns false",
			fullMsg:  "",
			prefix:   "!",
			wantCmd:  "",
			wantArgs: nil,
			wantOK:   false,
		},
		{
			name:     "only prefix returns false",
			fullMsg:  "!",
			prefix:   "!",
			wantCmd:  "",
			wantArgs: nil,
			wantOK:   false,
		},
		{
			name:     "prefix with only spaces returns false",
			fullMsg:  "!   ",
			prefix:   "!",
			wantCmd:  "",
			wantArgs: nil,
			wantOK:   false,
		},
		{
			name:     "extra spaces between args",
			fullMsg:  "!heist   1000   2000",
			prefix:   "!",
			wantCmd:  "heist",
			wantArgs: []string{"1000", "2000"},
			wantOK:   true,
		},
		{
			name:     "leading space after prefix",
			fullMsg:  "! status",
			prefix:   "!",
			wantCmd:  "status",
			wantArgs: nil,
			wantOK:   true,
		},
		{
			name:     "no prefix in message",
			fullMsg:  "status",
			prefix:   "!",
			wantCmd:  "",
			wantArgs: nil,
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotCmd, gotArgs, gotOK := splitCommand(tt.fullMsg, tt.prefix)

			assert.Equal(t, tt.wantOK, gotOK, "splitCommand() ok mismatch")
			assert.Equal(t, tt.wantCmd, gotCmd, "splitCommand() cmd mismatch")
			assert.Equal(t, tt.wantArgs, gotArgs, "splitCommand() args mismatch")
		})
	}
}
