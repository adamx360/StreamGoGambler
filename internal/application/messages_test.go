package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"streamgogambler/internal/adapters/logging"
	"streamgogambler/internal/ports"
)

func newTestMessageHandler() *MessageHandler {
	logger := logging.New(logging.LevelDebug)
	bot := &BotService{
		ctx: context.Background(),
	}
	return &MessageHandler{
		bot:    bot,
		logger: logger,
	}
}

func testConfig(username, prefix string) ports.BotConfig {
	return ports.BotConfig{
		Username: username,
		Prefix:   prefix,
	}
}

func TestIsCommandFromOwner(t *testing.T) {
	t.Parallel()

	logger := logging.New(logging.LevelDebug)
	h := &MessageHandler{logger: logger}

	tests := []struct {
		name     string
		username string
		message  string
		cfg      ports.BotConfig
		want     bool
	}{
		{
			name:     "owner with valid prefix",
			username: "testuser",
			message:  "!status",
			cfg:      testConfig("testuser", "!"),
			want:     true,
		},
		{
			name:     "owner case insensitive",
			username: "TestUser",
			message:  "!status",
			cfg:      testConfig("testuser", "!"),
			want:     true,
		},
		{
			name:     "owner without prefix",
			username: "testuser",
			message:  "status",
			cfg:      testConfig("testuser", "!"),
			want:     false,
		},
		{
			name:     "not owner with prefix",
			username: "otheruser",
			message:  "!status",
			cfg:      testConfig("testuser", "!"),
			want:     false,
		},
		{
			name:     "owner with different prefix",
			username: "testuser",
			message:  "?status",
			cfg:      testConfig("testuser", "?"),
			want:     true,
		},
		{
			name:     "owner wrong prefix",
			username: "testuser",
			message:  "?status",
			cfg:      testConfig("testuser", "!"),
			want:     false,
		},
		{
			name:     "empty username",
			username: "",
			message:  "!status",
			cfg:      testConfig("testuser", "!"),
			want:     false,
		},
		{
			name:     "empty message",
			username: "testuser",
			message:  "",
			cfg:      testConfig("testuser", "!"),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := h.isCommandFromOwner(tt.username, tt.message, tt.cfg)
			assert.Equal(t, tt.want, got, "isCommandFromOwner(%q, %q)", tt.username, tt.message)
		})
	}
}

func TestExtractTrustedUserCommand(t *testing.T) {
	t.Parallel()

	logger := logging.New(logging.LevelDebug)
	h := &MessageHandler{logger: logger}

	tests := []struct {
		name    string
		message string
		cfg     ports.BotConfig
		wantCmd string
		wantOK  bool
	}{
		{
			name:    "direct prefix command",
			message: "!status",
			cfg:     testConfig("botuser", "!"),
			wantCmd: "!status",
			wantOK:  true,
		},
		{
			name:    "direct prefix case insensitive",
			message: "!STATUS",
			cfg:     testConfig("botuser", "!"),
			wantCmd: "!status",
			wantOK:  true,
		},
		{
			name:    "@mention with prefix",
			message: "@botuser !status",
			cfg:     testConfig("botuser", "!"),
			wantCmd: "!status",
			wantOK:  true,
		},
		{
			name:    "@mention with comma and prefix",
			message: "@botuser, !status",
			cfg:     testConfig("botuser", "!"),
			wantCmd: "!status",
			wantOK:  true,
		},
		{
			name:    "@mention with command and args",
			message: "@botuser !ustaw 2000",
			cfg:     testConfig("botuser", "!"),
			wantCmd: "!ustaw 2000",
			wantOK:  true,
		},
		{
			name:    "@mention with comma command and args",
			message: "@botuser, !ustaw 2000",
			cfg:     testConfig("botuser", "!"),
			wantCmd: "!ustaw 2000",
			wantOK:  true,
		},
		{
			name:    "case insensitive @mention with command",
			message: "@Adamx360 !ustaw 2000",
			cfg:     testConfig("adamx360", "!"),
			wantCmd: "!ustaw 2000",
			wantOK:  true,
		},
		{
			name:    "username with prefix",
			message: "botuser !status",
			cfg:     testConfig("botuser", "!"),
			wantCmd: "!status",
			wantOK:  true,
		},
		{
			name:    "username with comma and prefix",
			message: "botuser, !status",
			cfg:     testConfig("botuser", "!"),
			wantCmd: "!status",
			wantOK:  true,
		},
		{
			name:    "no prefix at all",
			message: "hello world",
			cfg:     testConfig("botuser", "!"),
			wantCmd: "",
			wantOK:  false,
		},
		{
			name:    "@mention without prefix",
			message: "@botuser hello",
			cfg:     testConfig("botuser", "!"),
			wantCmd: "",
			wantOK:  false,
		},
		{
			name:    "username without prefix",
			message: "botuser hello",
			cfg:     testConfig("botuser", "!"),
			wantCmd: "",
			wantOK:  false,
		},
		{
			name:    "wrong user mention with prefix",
			message: "@otheruser !status",
			cfg:     testConfig("botuser", "!"),
			wantCmd: "",
			wantOK:  false,
		},
		{
			name:    "different prefix",
			message: "?help",
			cfg:     testConfig("botuser", "?"),
			wantCmd: "?help",
			wantOK:  true,
		},
		{
			name:    "case insensitive mention",
			message: "@BotUser !status",
			cfg:     testConfig("botuser", "!"),
			wantCmd: "!status",
			wantOK:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotCmd, gotOK := h.extractTrustedUserCommand(tt.message, tt.cfg)
			assert.Equal(t, tt.wantOK, gotOK, "extractTrustedUserCommand(%q) ok mismatch", tt.message)
			assert.Equal(t, tt.wantCmd, gotCmd, "extractTrustedUserCommand(%q) cmd mismatch", tt.message)
		})
	}
}

func TestDetectCooldown(t *testing.T) {
	t.Parallel()

	h := newTestMessageHandler()

	tests := []struct {
		name     string
		text     string
		username string
		want     bool
	}{
		{
			name:     "cooldown message",
			text:     "testuser is on cooldown for 30 seconds",
			username: "testuser",
			want:     true,
		},
		{
			name:     "cooldown case insensitive",
			text:     "TestUser is on COOLDOWN",
			username: "testuser",
			want:     true,
		},
		{
			name:     "no cooldown message",
			text:     "testuser has won 1000 points",
			username: "testuser",
			want:     false,
		},
		{
			name:     "cooldown for different user",
			text:     "otheruser is on cooldown",
			username: "testuser",
			want:     false,
		},
		{
			name:     "cooldown not at start",
			text:     "The user testuser has no cooldown",
			username: "testuser",
			want:     false,
		},
		{
			name:     "empty text",
			text:     "",
			username: "testuser",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := h.detectCooldown(tt.text, tt.username)
			assert.Equal(t, tt.want, got, "detectCooldown(%q, %q)", tt.text, tt.username)
		})
	}
}

func TestDetectInsufficientBombs(t *testing.T) {
	t.Parallel()

	h := newTestMessageHandler()

	tests := []struct {
		name     string
		text     string
		username string
		want     bool
	}{
		{
			name:     "insufficient bombs message",
			text:     "testuser doesn't have enough bombs to play",
			username: "testuser",
			want:     true,
		},
		{
			name:     "case insensitive",
			text:     "TestUser DOESN'T HAVE enough BOMBS",
			username: "testuser",
			want:     true,
		},
		{
			name:     "has bombs",
			text:     "testuser has 5000 bombs",
			username: "testuser",
			want:     false,
		},
		{
			name:     "different user insufficient",
			text:     "otheruser doesn't have enough bombs",
			username: "testuser",
			want:     false,
		},
		{
			name:     "partial match - only doesn't have",
			text:     "testuser doesn't have permission",
			username: "testuser",
			want:     false,
		},
		{
			name:     "partial match - only bombs",
			text:     "testuser has many bombs",
			username: "testuser",
			want:     false,
		},
		{
			name:     "empty text",
			text:     "",
			username: "testuser",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := h.detectInsufficientBombs(tt.text, tt.username)
			assert.Equal(t, tt.want, got, "detectInsufficientBombs(%q, %q)", tt.text, tt.username)
		})
	}
}
