package gambling

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateHeistAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		amount    int
		want      int
		wantError bool
	}{
		{"valid", 5000, 5000, false},
		{"min", 1, 1, false},
		{"max", MaxHeistAmount, MaxHeistAmount, false},
		{"over max clamped", MaxHeistAmount + 1000, MaxHeistAmount, false},
		{"zero invalid", 0, 0, true},
		{"negative invalid", -100, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ValidateHeistAmount(tt.amount)
			if tt.wantError {
				require.Error(t, err, "ValidateHeistAmount(%d) should return error", tt.amount)
			} else {
				require.NoError(t, err, "ValidateHeistAmount(%d) should not return error", tt.amount)
			}
			assert.Equal(t, tt.want, got, "ValidateHeistAmount(%d)", tt.amount)
		})
	}
}

func TestClampHeistAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		amount int
		want   int
	}{
		{"valid", 5000, 5000},
		{"min", 1, 1},
		{"max", MaxHeistAmount, MaxHeistAmount},
		{"over max", MaxHeistAmount + 1000, MaxHeistAmount},
		{"zero uses default", 0, DefaultHeistAmount},
		{"negative uses default", -100, DefaultHeistAmount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, ClampHeistAmount(tt.amount), "ClampHeistAmount(%d)", tt.amount)
		})
	}
}

func TestConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 10000, MaxHeistAmount, "MaxHeistAmount")
	assert.Equal(t, 1000, DefaultHeistAmount, "DefaultHeistAmount")
	assert.Equal(t, 2000, DefaultSlotsCost, "DefaultSlotsCost")
}
