package parsing

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const maxPointsStringLen = 20

const maxMessageLen = 1000

func ParsePoints(message, username string) (int, bool) {
	if strings.Contains(message, "top 5 users based on points") {
		return 0, false
	}

	if !utf8.ValidString(username) {
		return 0, false
	}

	if len(message) > maxMessageLen {
		return 0, false
	}

	pattern := `^.*` + regexp.QuoteMeta(username) + `\s*\(\s*([\d\s]+)\).*`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return 0, false
	}

	matches := re.FindStringSubmatch(message)
	if len(matches) != 2 {
		return 0, false
	}

	if len(matches[1]) > maxPointsStringLen {
		return 0, false
	}

	pointsStr := strings.ReplaceAll(matches[1], " ", "")

	n, err := strconv.Atoi(pointsStr)
	if err != nil {
		return 0, false
	}

	return n, true
}
