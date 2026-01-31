package wallet

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initialBalance int
		wantBalance    int
	}{
		{"zero balance", 0, 0},
		{"positive balance", 1000, 1000},
		{"large balance", 1000000, 1000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := New(tt.initialBalance)
			assert.Equal(t, tt.wantBalance, w.GetBalance(), "New(%d).GetBalance()", tt.initialBalance)
		})
	}
}

func TestWallet_SetBalance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		initial     int
		setTo       int
		wantBalance int
	}{
		{"set to zero", 1000, 0, 0},
		{"set to positive", 0, 2000, 2000},
		{"set to same", 1000, 1000, 1000},
		{"set to larger", 500, 1500, 1500},
		{"set to smaller", 1500, 500, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := New(tt.initial)
			w.SetBalance(tt.setTo)
			assert.Equal(t, tt.wantBalance, w.GetBalance(), "SetBalance(%d) → GetBalance()", tt.setTo)
		})
	}
}

func TestWallet_AddBalance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		initial     int
		add         int
		wantBalance int
	}{
		{"add positive", 1000, 500, 1500},
		{"add zero", 1000, 0, 1000},
		{"add negative (subtract)", 1000, -200, 800},
		{"add to zero", 0, 500, 500},
		{"subtract to zero", 500, -500, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := New(tt.initial)
			w.AddBalance(tt.add)
			assert.Equal(t, tt.wantBalance, w.GetBalance(), "AddBalance(%d) → GetBalance()", tt.add)
		})
	}
}

func TestWallet_Spend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		initial     int
		spend       int
		wantSuccess bool
		wantBalance int
	}{
		{"spend some", 1000, 500, true, 500},
		{"spend exact", 1000, 1000, true, 0},
		{"spend zero", 1000, 0, true, 1000},
		{"spend more than balance", 1000, 1500, false, 1000},
		{"spend from zero", 0, 100, false, 0},
		{"spend zero from zero", 0, 0, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := New(tt.initial)
			got := w.Spend(tt.spend)
			assert.Equal(t, tt.wantSuccess, got, "Spend(%d) success", tt.spend)
			assert.Equal(t, tt.wantBalance, w.GetBalance(), "Spend(%d) → GetBalance()", tt.spend)
		})
	}
}

func TestWallet_CanAfford(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		initial int
		amount  int
		want    bool
	}{
		{"can afford less", 1000, 500, true},
		{"can afford exact", 1000, 1000, true},
		{"cannot afford more", 1000, 1001, false},
		{"can afford zero", 1000, 0, true},
		{"zero balance afford zero", 0, 0, true},
		{"zero balance cannot afford", 0, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := New(tt.initial)
			assert.Equal(t, tt.want, w.CanAfford(tt.amount), "CanAfford(%d)", tt.amount)
		})
	}
}

func TestWallet_Concurrent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		initial     int
		goroutines  int
		spendEach   int
		wantBalance int
	}{
		{"100 goroutines spend 100 each", 10000, 100, 100, 0},
		{"50 goroutines spend 200 each", 10000, 50, 200, 0},
		{"10 goroutines spend 500 each", 5000, 10, 500, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := New(tt.initial)
			var wg sync.WaitGroup

			for i := 0; i < tt.goroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					w.Spend(tt.spendEach)
				}()
			}

			wg.Wait()

			assert.Equal(t, tt.wantBalance, w.GetBalance(), "After concurrent spend: GetBalance()")
		})
	}
}
