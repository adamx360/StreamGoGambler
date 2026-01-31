package gambling

import "errors"

const (
	MaxHeistAmount = 10000

	DefaultHeistAmount = 1000

	DefaultSlotsCost = 2000

	DefaultArenaCost = 1000
)

var ErrInvalidAmount = errors.New("invalid heist amount")

func ValidateHeistAmount(amount int) (int, error) {
	if amount <= 0 {
		return 0, ErrInvalidAmount
	}
	if amount > MaxHeistAmount {
		return MaxHeistAmount, nil // Clamp to max
	}
	return amount, nil
}

func ClampHeistAmount(amount int) int {
	if amount <= 0 {
		return DefaultHeistAmount
	}
	if amount > MaxHeistAmount {
		return MaxHeistAmount
	}
	return amount
}
