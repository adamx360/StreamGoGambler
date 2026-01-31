package parsing

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBombs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		message  string
		username string
		want     int
		wantOK   bool
	}{
		{"valid", "testuser bombs: 1234", "testuser", 1234, true},
		{"case insensitive", "TestUser Bombs: 5000", "testuser", 5000, true},
		{"wrong user", "otheruser bombs: 1234", "testuser", 0, false},
		{"no bombs", "testuser has no bombs", "testuser", 0, false},
		{"empty message", "", "testuser", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := ParseBombs(tt.message, tt.username)
			assert.Equal(t, tt.wantOK, ok, "ParseBombs() ok")
			assert.Equal(t, tt.want, got, "ParseBombs() value")
		})
	}
}

func TestParseSlotsDelta(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		message     string
		username    string
		wantDelta   int
		wantOutcome SlotsOutcome
		wantOK      bool
	}{
		{"lost", "testuser you lost everything", "testuser", 0, OutcomeLost, true},
		{"refund", "testuser you at least got your points back", "testuser", 2000, OutcomeRefund, true},
		{"small win", "testuser even a small win is a win", "testuser", 6000, OutcomeSmallWin, true},
		{"jackpot", "testuser hit the jackpot!", "testuser", 20000, OutcomeJackpot, true},
		{"super jackpot", "testuser hit the SUPER JACKPOT!", "testuser", 60000, OutcomeSuperJackpot, true},
		{"wrong user", "otheruser you lost", "testuser", 0, "", false},
		{"unknown", "testuser something random", "testuser", 0, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, ok := ParseSlotsDelta(tt.message, tt.username)
			assert.Equal(t, tt.wantOK, ok, "ParseSlotsDelta() ok")
			assert.Equal(t, tt.wantDelta, result.Delta, "ParseSlotsDelta() delta")
			assert.Equal(t, tt.wantOutcome, result.Outcome, "ParseSlotsDelta() outcome")
		})
	}
}

func TestParsePoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		message  string
		username string
		want     int
		wantOK   bool
	}{
		{"simple", "testuser (1234)", "testuser", 1234, true},
		{"with spaces", "testuser (2 000)", "testuser", 2000, true},
		{"large number", "testuser (1 234 567)", "testuser", 1234567, true},
		{"in context", "Results: testuser (500) won", "testuser", 500, true},
		{"wrong user", "otheruser (1234)", "testuser", 0, false},
		{"no parens", "testuser 1234", "testuser", 0, false},
		{"top 5 filtered", "top 5 users based on points: testuser (1000)", "testuser", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := ParsePoints(tt.message, tt.username)
			assert.Equal(t, tt.wantOK, ok, "ParsePoints() ok")
			assert.Equal(t, tt.want, got, "ParsePoints() value")
		})
	}
}

func TestParsePointsMaxLength(t *testing.T) {
	t.Parallel()

	longMsg := "testuser (" + strings.Repeat("1", 1001) + ")"
	_, ok := ParsePoints(longMsg, "testuser")
	assert.False(t, ok, "ParsePoints should reject messages over 1000 chars")
}

func TestParsePointsInvalidUTF8(t *testing.T) {
	t.Parallel()

	_, ok := ParsePoints("test (1000)", "\xff\xfe")
	assert.False(t, ok, "ParsePoints should reject invalid UTF-8 usernames")
}

func FuzzParseBombs(f *testing.F) {
	f.Add("testuser bombs: 1234", "testuser")
	f.Add("", "")
	f.Add("random message", "user")

	f.Fuzz(func(t *testing.T, message, username string) {
		ParseBombs(message, username)
	})
}

func FuzzParseSlotsDelta(f *testing.F) {
	f.Add("testuser you lost", "testuser")
	f.Add("testuser jackpot!", "testuser")
	f.Add("", "")

	f.Fuzz(func(t *testing.T, message, username string) {
		ParseSlotsDelta(message, username)
	})
}

func FuzzParsePoints(f *testing.F) {
	f.Add("testuser (1234)", "testuser")
	f.Add("user (2 000)", "user")
	f.Add("", "")

	f.Fuzz(func(t *testing.T, message, username string) {
		ParsePoints(message, username)
	})
}
