package application

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"streamgogambler/internal/adapters/logging"
	"streamgogambler/internal/domain/gambling"
	"streamgogambler/internal/ports"
)

type CommandHandler struct {
	bot    *BotService
	config ports.ConfigStore
	logger *logging.Logger
	cmds   map[string]CommandFunc
}

type CommandFunc func(userName, channel string, args []string)

func NewCommandHandler(bot *BotService, config ports.ConfigStore, logger *logging.Logger) *CommandHandler {
	h := &CommandHandler{
		bot:    bot,
		config: config,
		logger: logger,
		cmds:   make(map[string]CommandFunc),
	}

	cfg := config.GetConfig()
	h.cmds[strings.ToLower(cfg.StatusCommand)] = h.handleStatus
	h.cmds["ustaw"] = h.handleSetHeist
	h.cmds["jakiheist"] = h.handleCheckHeist
	h.cmds["autoslots"] = h.handleAutoSlots
	h.cmds["slotsoff"] = h.handleSlotsOff
	h.cmds["trust"] = h.handleTrust
	h.cmds["untrust"] = h.handleUntrust
	h.cmds["trustlist"] = h.handleTrustList
	h.cmds["help"] = h.handleHelp

	return h
}

func (h *CommandHandler) HandleCommand(userName, channel, fullMsg string) {
	cfg := h.config.GetConfig()

	cmd, args, ok := splitCommand(fullMsg, cfg.Prefix)
	if !ok {
		return
	}

	isOwner := strings.EqualFold(userName, cfg.Username)
	if !isOwner && h.bot.IsUserRateLimited(userName) {
		h.logger.Debugf(h.bot.ctx, "Command blocked (rate limit) from %s: %s", userName, cmd)
		return
	}

	if handler, ok := h.cmds[cmd]; ok {
		handler(userName, channel, args)
	}
}

func (h *CommandHandler) IsInternalCommand(cmdName string) bool {
	_, exists := h.cmds[strings.ToLower(cmdName)]
	return exists
}

func (h *CommandHandler) handleStatus(userName, channel string, _ []string) {
	if !h.bot.IsUserTrusted(userName) {
		return
	}

	cfg := h.config.GetConfig()
	msg := fmt.Sprintf("@%s, Bot działa prawidłowo ;) | Bombs: %d | Heist: %d",
		userName, h.bot.Wallet().GetBalance(), cfg.DefaultHeist)
	h.bot.SafeSay(channel, msg)
}

func (h *CommandHandler) handleSetHeist(userName, channel string, args []string) {
	if !h.bot.IsUserTrusted(userName) {
		return
	}

	if len(args) == 0 {
		h.bot.SafeSay(channel, fmt.Sprintf("@%s, Podaj liczbę od 1 do max %d!", userName, gambling.MaxHeistAmount))
		return
	}

	heist, err := strconv.Atoi(args[0])
	if err != nil || heist <= 0 || heist > gambling.MaxHeistAmount {
		h.bot.SafeSay(channel, fmt.Sprintf("@%s, Podaj liczbę od 1 do max %d!", userName, gambling.MaxHeistAmount))
		return
	}

	if err := h.config.UpdateHeist(heist); err != nil {
		h.logger.Errorf(h.bot.ctx, "Error updating HEIST_AMOUNT in .env: %v", err)
		h.bot.SafeSay(channel, fmt.Sprintf("@%s, Wystąpił błąd podczas aktualizacji wartości heist!", userName))
		return
	}

	h.bot.SafeSay(channel, fmt.Sprintf("@%s, Pomyślnie zmieniono ilość heista na %d!", userName, heist))
	h.logger.Infof(h.bot.ctx, "Successfully updated HEIST_AMOUNT to %d", heist)
}

func (h *CommandHandler) handleCheckHeist(userName, channel string, _ []string) {
	if !h.bot.IsUserTrusted(userName) {
		return
	}

	cfg := h.config.GetConfig()
	h.bot.SafeSay(channel, fmt.Sprintf("@%s, Masz aktualnie ustawione %d heista ;)", userName, cfg.DefaultHeist))
}

func (h *CommandHandler) handleAutoSlots(userName, channel string, args []string) {
	if !h.bot.IsUserTrusted(userName) {
		return
	}

	if len(args) == 0 {
		status := "wyłączone"
		if h.bot.IsAutoSlotsEnabled() {
			status = "włączone"
		}
		h.bot.SafeSay(channel, fmt.Sprintf("@%s, Auto slots jest %s. Użyj: !autoslots on/off", userName, status))
		return
	}

	switch strings.ToLower(args[0]) {
	case "on", "1", "true", "wlacz", "włącz":
		h.bot.SetAutoSlots(true)
		h.bot.SafeSay(channel, fmt.Sprintf("@%s, Auto slots włączone!", userName))
		h.logger.Infof(h.bot.ctx, "Auto slots enabled by %s", userName)
	case "off", "0", "false", "wylacz", "wyłącz":
		h.bot.SetAutoSlots(false)
		h.bot.SafeSay(channel, fmt.Sprintf("@%s, Auto slots wyłączone!", userName))
		h.logger.Infof(h.bot.ctx, "Auto slots disabled by %s", userName)
	default:
		h.bot.SafeSay(channel, fmt.Sprintf("@%s, Użyj: !autoslots on/off", userName))
	}
}

func (h *CommandHandler) handleSlotsOff(userName, channel string, args []string) {
	if !h.bot.IsUserTrusted(userName) {
		return
	}

	if len(args) == 0 {
		offTime := h.bot.GetSlotsOffTime()
		if offTime.IsZero() {
			h.bot.SafeSay(channel, fmt.Sprintf("@%s, Brak zaplanowanego wyłączenia. Użyj: !slotsoff <czas> lub !slotsoff <duration>", userName))
		} else {
			remaining := time.Until(offTime).Round(time.Second)
			h.bot.SafeSay(channel, fmt.Sprintf("@%s, Auto slots wyłączy się o %s (za %s)", userName, offTime.Format("15:04"), remaining))
		}
		return
	}

	arg := strings.ToLower(args[0])

	if arg == "cancel" || arg == "anuluj" {
		if h.bot.CancelSlotsOffSchedule() {
			h.bot.SafeSay(channel, fmt.Sprintf("@%s, Anulowano zaplanowane wyłączenie.", userName))
			h.logger.Infof(h.bot.ctx, "Slots off schedule canceled by %s", userName)
		} else {
			h.bot.SafeSay(channel, fmt.Sprintf("@%s, Brak zaplanowanego wyłączenia.", userName))
		}
		return
	}

	if matched, _ := regexp.MatchString(`^\d{1,2}:\d{2}$`, arg); matched {
		parts := strings.Split(arg, ":")
		hour, _ := strconv.Atoi(parts[0])
		minute, _ := strconv.Atoi(parts[1])

		if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
			h.bot.SafeSay(channel, fmt.Sprintf("@%s, Nieprawidłowy czas. Użyj formatu HH:MM (np. 22:00)", userName))
			return
		}

		now := time.Now()
		offTime := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

		if offTime.Before(now) {
			offTime = offTime.Add(24 * time.Hour)
		}

		h.bot.ScheduleSlotsOff(offTime)
		h.bot.SafeSay(channel, fmt.Sprintf("@%s, Auto slots wyłączy się o %s", userName, offTime.Format("15:04")))
		h.logger.Infof(h.bot.ctx, "Slots off scheduled for %s by %s", offTime.Format("15:04"), userName)
		return
	}

	duration, err := time.ParseDuration(arg)
	if err != nil || duration <= 0 {
		h.bot.SafeSay(channel, fmt.Sprintf("@%s, Użyj: !slotsoff <HH:MM> lub !slotsoff <duration> (np. 2h, 30m, 1h30m)", userName))
		return
	}

	offTime := time.Now().Add(duration)
	h.bot.ScheduleSlotsOff(offTime)
	h.bot.SafeSay(channel, fmt.Sprintf("@%s, Auto slots wyłączy się za %s (o %s)", userName, duration.Round(time.Second), offTime.Format("15:04")))
	h.logger.Infof(h.bot.ctx, "Slots off scheduled in %s by %s", duration, userName)
}

func (h *CommandHandler) handleTrust(userName, channel string, args []string) {
	cfg := h.config.GetConfig()
	if !strings.EqualFold(userName, cfg.Username) {
		return
	}

	if len(args) == 0 {
		h.bot.SafeSay(channel, fmt.Sprintf("@%s, Użyj: !trust <nick>", userName))
		return
	}

	target := strings.ToLower(args[0])
	if strings.EqualFold(target, cfg.Username) {
		h.bot.SafeSay(channel, fmt.Sprintf("@%s, Nie możesz dodać siebie do listy!", userName))
		return
	}

	if h.bot.IsUserTrusted(target) {
		h.bot.SafeSay(channel, fmt.Sprintf("@%s, %s już jest zaufanym użytkownikiem", userName, target))
		return
	}

	h.bot.AddTrustedUser(target)
	h.bot.SafeSay(channel, fmt.Sprintf("@%s, Dodano %s do zaufanych użytkowników!", userName, target))
	h.logger.Infof(h.bot.ctx, "Added %s to trusted users by %s", target, userName)
}

func (h *CommandHandler) handleUntrust(userName, channel string, args []string) {
	cfg := h.config.GetConfig()
	if !strings.EqualFold(userName, cfg.Username) {
		return
	}

	if len(args) == 0 {
		h.bot.SafeSay(channel, fmt.Sprintf("@%s, Użyj: !untrust <nick>", userName))
		return
	}

	target := strings.ToLower(args[0])
	if !h.bot.IsUserTrusted(target) {
		h.bot.SafeSay(channel, fmt.Sprintf("@%s, %s nie jest zaufanym użytkownikiem", userName, target))
		return
	}

	h.bot.RemoveTrustedUser(target)
	h.bot.SafeSay(channel, fmt.Sprintf("@%s, Usunięto %s z zaufanych użytkowników!", userName, target))
	h.logger.Infof(h.bot.ctx, "Removed %s from trusted users by %s", target, userName)
}

func (h *CommandHandler) handleTrustList(userName, channel string, _ []string) {
	cfg := h.config.GetConfig()
	if !strings.EqualFold(userName, cfg.Username) {
		return
	}

	users := h.bot.GetTrustedUsers()
	if len(users) == 0 {
		h.bot.SafeSay(channel, fmt.Sprintf("@%s, Lista zaufanych jest pusta.", userName))
		return
	}

	h.bot.SafeSay(channel, fmt.Sprintf("@%s, Zaufani: %s", userName, strings.Join(users, ", ")))
}

func (h *CommandHandler) handleHelp(userName, channel string, _ []string) {
	if !h.bot.IsUserTrusted(userName) {
		return
	}

	cfg := h.config.GetConfig()
	help := strings.Join([]string{
		"Komendy (zaufani):",
		fmt.Sprintf("%s%s — status bota", cfg.Prefix, strings.ToLower(cfg.StatusCommand)),
		fmt.Sprintf("%sustaw <kwota> — ustawia heist", cfg.Prefix),
		fmt.Sprintf("%sjakiheist — pokazuje heist", cfg.Prefix),
		fmt.Sprintf("%sautoslots on/off — auto slots", cfg.Prefix),
		fmt.Sprintf("%sslotsoff <czas/duration> — planuje wyłączenie", cfg.Prefix),
		fmt.Sprintf("%shelp — ta pomoc", cfg.Prefix),
	}, " | ")
	h.bot.SafeSay(channel, help)
}

func splitCommand(fullMsg, prefix string) (string, []string, bool) {
	if !strings.HasPrefix(fullMsg, prefix) {
		return "", nil, false
	}
	text := strings.TrimSpace(strings.TrimPrefix(fullMsg, prefix))
	if text == "" {
		return "", nil, false
	}
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return "", nil, false
	}
	cmd := strings.ToLower(parts[0])
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}
	return cmd, args, true
}

var _ = context.Background
