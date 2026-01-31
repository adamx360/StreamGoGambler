package parsing

import "strings"

type SlotsOutcome string

const (
	OutcomeLost         SlotsOutcome = "lost"
	OutcomeRefund       SlotsOutcome = "refund"
	OutcomeSmallWin     SlotsOutcome = "small_win"
	OutcomeJackpot      SlotsOutcome = "jackpot"
	OutcomeSuperJackpot SlotsOutcome = "super_jackpot"
)

type SlotsResult struct {
	Delta   int
	Outcome SlotsOutcome
}

func ParseSlotsDelta(message, username string) (SlotsResult, bool) {
	lower := strings.ToLower(message)
	if !strings.Contains(lower, strings.ToLower(username)) {
		return SlotsResult{}, false
	}

	switch {
	case strings.Contains(lower, "you lost"):
		return SlotsResult{Delta: 0, Outcome: OutcomeLost}, true

	case strings.Contains(lower, "super jackpot"):
		return SlotsResult{Delta: 60000, Outcome: OutcomeSuperJackpot}, true

	case strings.Contains(lower, "jackpot"):
		return SlotsResult{Delta: 20000, Outcome: OutcomeJackpot}, true

	case strings.Contains(lower, "even a small win is a win"):
		return SlotsResult{Delta: 6000, Outcome: OutcomeSmallWin}, true

	case strings.Contains(lower, "you at least got your points back"):
		return SlotsResult{Delta: 2000, Outcome: OutcomeRefund}, true
	case strings.Contains(lower, "he command is still on user cooldown for"):
		return SlotsResult{Delta: 2000, Outcome: OutcomeRefund}, true
	default:
		return SlotsResult{}, false
	}
}
