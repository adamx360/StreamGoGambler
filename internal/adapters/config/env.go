package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"streamgogambler/internal/domain/gambling"
	"streamgogambler/internal/ports"
)

const (
	DefaultSayBucketSize = 20
	DefaultSayRefillMs   = 150
	trueString           = "true"
)

type EnvStore struct {
	envPath string
	config  ports.BotConfig
	oauth   string
}

func NewEnvStore(envPath string) (*EnvStore, error) {
	store := &EnvStore{envPath: envPath}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *EnvStore) load() error {
	required := []string{
		"TWITCH_USERNAME",
		"TWITCH_OAUTH",
		"TWITCH_CHANNEL",
		"COMMAND_PREFIX",
		"STATUS_COMMAND",
		"CONNECT_MESSAGE",
		"BOSS_BOT_NAME",
	}

	var missing []string
	for _, k := range required {
		if os.Getenv(k) == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	heist, _ := strconv.Atoi(getEnv("HEIST_AMOUNT", strconv.Itoa(gambling.DefaultHeistAmount)))
	slotsCost, _ := strconv.Atoi(getEnv("SLOTS_COST", strconv.Itoa(gambling.DefaultSlotsCost)))
	arenaCost, _ := strconv.Atoi(getEnv("ARENA_COST", strconv.Itoa(gambling.DefaultArenaCost)))
	autoSlotsEnabled := strings.ToLower(getEnv("AUTO_SLOTS_ENABLED", "false")) == trueString
	autoSlotsInterval, _ := strconv.Atoi(getEnv("AUTO_SLOTS_INTERVAL", "15"))
	bandOnPerma := strings.ToLower(getEnv("BAND_ON_PERMA", "false")) == trueString
	pointsAsDelta := strings.ToLower(getEnv("POINTS_AS_DELTA", trueString)) == trueString
	bucketSize, _ := strconv.Atoi(getEnv("SAY_BUCKET_SIZE", strconv.Itoa(DefaultSayBucketSize)))
	refillMs, _ := strconv.Atoi(getEnv("SAY_REFILL_MS", strconv.Itoa(DefaultSayRefillMs)))
	greetOnReconnect := strings.ToLower(getEnv("GREET_ON_RECONNECT", "false")) == trueString
	healthPort, _ := strconv.Atoi(getEnv("HEALTH_PORT", "0"))
	guiEnabled := strings.ToLower(getEnv("GUI_ENABLED", trueString)) == trueString
	maxLogsLines, _ := strconv.Atoi(getEnv("MAX_LOGS_LINES", "500"))

	autoResponses := map[string]string{
		"Type !boss to start!": "!boss",
		"Type !boss to join!":  "!boss",
		"Type !ffa to start!":  "!ffa",
		"!los":                 "!los",
		"The cops have given up! If you want to get a team together type !heist": "!heist",
	}

	s.config = ports.BotConfig{
		Username:          os.Getenv("TWITCH_USERNAME"),
		Channel:           os.Getenv("TWITCH_CHANNEL"),
		Prefix:            os.Getenv("COMMAND_PREFIX"),
		StatusCommand:     os.Getenv("STATUS_COMMAND"),
		ConnectMessage:    os.Getenv("CONNECT_MESSAGE"),
		BossBotName:       os.Getenv("BOSS_BOT_NAME"),
		DefaultHeist:      heist,
		SlotsCost:         slotsCost,
		ArenaCost:         arenaCost,
		AutoSlotsEnabled:  autoSlotsEnabled,
		AutoSlotsInterval: autoSlotsInterval,
		BandOnPerma:       bandOnPerma,
		BandMessage:       getEnv("BAND_MESSAGE", "BAND"),
		PointsAsDelta:     pointsAsDelta,
		SayBucketSize:     bucketSize,
		SayRefillMs:       refillMs,
		GreetOnReconnect:  greetOnReconnect,
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		HealthPort:        healthPort,
		AutoResponses:     autoResponses,
		GUIEnabled:        guiEnabled,
		MaxLogsLines:      maxLogsLines,
	}

	s.oauth = os.Getenv("TWITCH_OAUTH")
	return nil
}

func (s *EnvStore) GetConfig() ports.BotConfig {
	return s.config
}

func (s *EnvStore) GetOAuth() string {
	return s.oauth
}

func (s *EnvStore) UpdateHeist(amount int) error {
	if err := updateEnvFile(s.envPath, "HEIST_AMOUNT", strconv.Itoa(amount)); err != nil {
		return err
	}
	s.config.DefaultHeist = amount
	return nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func updateEnvFile(envPath, key, value string) error {
	envPath = filepath.Clean(envPath)
	dir := filepath.Dir(envPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	var lines []string
	// #nosec G304 -- envPath is intentionally user-configurable via ENV_PATH
	if f, err := os.Open(envPath); err == nil {
		scanner := bufio.NewScanner(f)
		found := false
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, key+"=") {
				lines = append(lines, fmt.Sprintf("%s=%s", key, value))
				found = true
			} else {
				lines = append(lines, line)
			}
		}
		_ = f.Close()
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading %s: %w", envPath, err)
		}
		if !found {
			lines = append(lines, fmt.Sprintf("%s=%s", key, value))
		}
	} else {
		lines = []string{fmt.Sprintf("%s=%s", key, value)}
	}

	content := strings.Join(lines, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	tmp, err := os.CreateTemp(dir, ".env.tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()
	if _, err := tmp.WriteString(content); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("writing to temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpPath, envPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("renaming %s to %s: %w", tmpPath, envPath, err)
	}

	_ = os.Setenv(key, value)
	return nil
}

func ResolveEnvPath() string {
	if p := os.Getenv("ENV_PATH"); p != "" {
		return p
	}
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), ".env")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ".env"
	}
	return filepath.Join(cwd, ".env")
}

func RequiredVariables() []string {
	return []string{
		"TWITCH_USERNAME",
		"TWITCH_OAUTH",
		"TWITCH_CHANNEL",
		"COMMAND_PREFIX",
		"STATUS_COMMAND",
		"CONNECT_MESSAGE",
		"BOSS_BOT_NAME",
	}
}

func GetMissingVariables() []string {
	var missing []string
	for _, k := range RequiredVariables() {
		if os.Getenv(k) == "" {
			missing = append(missing, k)
		}
	}
	return missing
}

func SaveConfigToEnv(envPath string, values map[string]string) error {
	for key, value := range values {
		if err := updateEnvFile(envPath, key, value); err != nil {
			return fmt.Errorf("failed to save %s: %w", key, err)
		}
		_ = os.Setenv(key, value)
	}
	return nil
}

func GetDefaultValues() map[string]string {
	return map[string]string{
		"COMMAND_PREFIX":  "!",
		"STATUS_COMMAND":  "status",
		"CONNECT_MESSAGE": "!pyk",
		"BOSS_BOT_NAME":   "demonzzbot",
		"GUI_ENABLED":     trueString,
	}
}
