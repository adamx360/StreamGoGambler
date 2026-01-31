package parsing

import (
	"regexp"
	"strconv"
	"strings"
)

var bombsRe = regexp.MustCompile(`(?i)\bbombs:\s*(\d+)`)

func ParseBombs(message, username string) (int, bool) {
	lower := strings.ToLower(message)
	if !strings.Contains(lower, strings.ToLower(username)) {
		return 0, false
	}

	m := bombsRe.FindStringSubmatch(message)
	if len(m) != 2 {
		return 0, false
	}

	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, false
	}

	return n, true
}
