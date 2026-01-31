package application

import (
	"context"
	"fmt"
	"strings"

	"streamgogambler/internal/adapters/logging"
	"streamgogambler/internal/domain/parsing"
	"streamgogambler/internal/ports"
)

type MessageHandler struct {
	bot    *BotService
	logger *logging.Logger
}

func NewMessageHandler(bot *BotService, logger *logging.Logger) *MessageHandler {
	return &MessageHandler{
		bot:    bot,
		logger: logger,
	}
}

func (h *MessageHandler) HandleMessage(msg ports.ChatMessage) {
	cfg := h.bot.Config().GetConfig()

	if strings.EqualFold(msg.UserName, cfg.BossBotName) {
		h.handleTrustedBotMessage(msg, cfg)
		return
	}
	message := strings.ToLower(msg.Text)
	if h.isCommandFromOwner(msg.UserName, message, cfg) {
		h.bot.cmdHandler.HandleCommand(msg.UserName, msg.Channel, message)
	} else if h.bot.IsUserTrusted(msg.UserName) {
		if cmd, ok := h.extractTrustedUserCommand(message, cfg); ok {
			h.bot.cmdHandler.HandleCommand(msg.UserName, msg.Channel, cmd)
		}
	}
}

func (h *MessageHandler) handleTrustedBotMessage(msg ports.ChatMessage, cfg ports.BotConfig) {
	text := h.handleSplitMessage(msg.Text)
	lower := strings.ToLower(text)

	if strings.Contains(lower, "bombs:") && strings.Contains(lower, strings.ToLower(cfg.Username)) {
		h.handleBombsResponse(text, cfg.Username)
		return
	}

	if strings.HasPrefix(text, cfg.Username+" pulls the lever and waits for the roll") {
		h.handleSlotsResponse(text, cfg.Username)
		return
	}

	if strings.Contains(text, cfg.Username+" (") {
		if strings.Contains(text, "Results from the Heist:") {
			h.handleHeistResult(text, cfg.Username)
			return
		}
		if strings.Contains(text, "The dust finally settled") {
			h.handleArenaResult(text, cfg.Username)
			return
		}
		h.handlePointsResponse(text, cfg)
		return
	}

	if h.checkSplitMessageStart(text, cfg.Username) {
		return
	}

	h.handleBossBotPrompts(text, msg.Channel, cfg)
	h.logger.Debugf(h.bot.ctx, "[%d] #%s %s -> %s", h.bot.Wallet().GetBalance(), msg.Channel, msg.UserName, text)
}

func (h *MessageHandler) isCommandFromOwner(username string, message string, cfg ports.BotConfig) bool {
	if !strings.EqualFold(username, cfg.Username) {
		return false
	}

	return strings.HasPrefix(message, strings.ToLower(cfg.Prefix))
}

func (h *MessageHandler) extractTrustedUserCommand(message string, cfg ports.BotConfig) (string, bool) {
	lower := strings.ToLower(message)
	usernamePrefix := strings.ToLower(cfg.Username)
	prefixLower := strings.ToLower(cfg.Prefix)

	if strings.HasPrefix(lower, prefixLower) {
		return lower, true
	}

	atMention := fmt.Sprintf("@%s", usernamePrefix)
	if strings.HasPrefix(lower, atMention) {
		rest := strings.TrimPrefix(lower, atMention)
		rest = strings.TrimSpace(strings.TrimPrefix(rest, ","))
		rest = strings.TrimSpace(rest)
		if strings.HasPrefix(rest, prefixLower) {
			return rest, true
		}
	}

	if strings.HasPrefix(lower, usernamePrefix) {
		rest := strings.TrimPrefix(lower, usernamePrefix)
		rest = strings.TrimSpace(strings.TrimPrefix(rest, ","))
		rest = strings.TrimSpace(rest)
		if strings.HasPrefix(rest, prefixLower) {
			return rest, true
		}
	}

	return "", false
}

func (h *MessageHandler) handleBombsResponse(text, username string) {
	if count, ok := parsing.ParseBombs(text, username); ok {
		h.bot.Wallet().SetBalance(count)
		h.logger.Infof(h.bot.ctx, "Updated bombs for %s: %d", username, count)
	} else {
		h.logger.Debugf(h.bot.ctx, "Could not parse bombs from: %s", text)
	}
}

func (h *MessageHandler) handleSlotsResponse(text, username string) {
	if result, ok := parsing.ParseSlotsDelta(text, username); ok {
		if result.Delta != 0 {
			h.bot.Wallet().AddBalance(result.Delta)
		}
		h.bot.RecordSlotsPlayed()
		h.logger.Infof(h.bot.ctx, "Slots result: %s | Bombs: %d", result.Outcome, h.bot.Wallet().GetBalance())
	} else {
		h.logger.Debugf(h.bot.ctx, "Unknown slots result: %s", text)
	}
}

func (h *MessageHandler) handlePointsResponse(text string, cfg ports.BotConfig) {
	if points, ok := parsing.ParsePoints(text, cfg.Username); ok {
		old := h.bot.Wallet().GetBalance()
		if cfg.PointsAsDelta {
			h.bot.Wallet().AddBalance(points)
			h.logger.Infof(h.bot.ctx, "+%d points → Bombs: %d → %d", points, old, h.bot.Wallet().GetBalance())
		} else {
			h.bot.Wallet().SetBalance(points)
			h.logger.Infof(h.bot.ctx, "Set bombs to %d (from points)", points)
		}
	} else {
		h.logger.Debugf(h.bot.ctx, "Could not parse points from: %s", text)
	}
}

func (h *MessageHandler) handleHeistResult(text, username string) {
	if payout, ok := parsing.ParsePoints(text, username); ok {
		old := h.bot.Wallet().GetBalance()
		h.bot.Wallet().AddBalance(payout)
		h.logger.Infof(h.bot.ctx, "Heist finished! Won: %d | Bombs: %d → %d", payout, old, h.bot.Wallet().GetBalance())
	} else {
		h.logger.Debugf(h.bot.ctx, "Could not parse heist payout from: %s", text)
	}
}

func (h *MessageHandler) handleArenaResult(text, username string) {
	if payout, ok := parsing.ParsePoints(text, username); ok {
		old := h.bot.Wallet().GetBalance()
		h.bot.Wallet().AddBalance(payout)
		h.logger.Infof(h.bot.ctx, "Arena finished! Won: %d | Bombs: %d → %d", payout, old, h.bot.Wallet().GetBalance())
	} else {
		h.logger.Debugf(h.bot.ctx, "Could not parse arena payout from: %s", text)
	}
}

func (h *MessageHandler) handleBossBotPrompts(text, channel string, cfg ports.BotConfig) {
	if h.detectCooldown(text, cfg.Username) {
		return
	}

	if h.detectInsufficientBombs(text, cfg.Username) {
		return
	}

	for trigger, response := range cfg.AutoResponses {
		if strings.Contains(text, trigger) {
			if response == "!heist" {
				response = fmt.Sprintf("!heist %d", cfg.DefaultHeist)
			}
			h.bot.SafeSay(channel, response)
			break
		}
	}
}

func (h *MessageHandler) detectCooldown(text, username string) bool {
	lower := strings.ToLower(text)
	userLower := strings.ToLower(username)

	if strings.HasPrefix(lower, userLower) && strings.Contains(lower, "cooldown") {
		h.logger.Warnf(h.bot.ctx, "Bot is on cooldown: %s", text)
		return true
	}
	return false
}

func (h *MessageHandler) detectInsufficientBombs(text, username string) bool {
	lower := strings.ToLower(text)
	userLower := strings.ToLower(username)

	if strings.Contains(lower, userLower) && strings.Contains(lower, "doesn't have") && strings.Contains(lower, "bombs") {
		h.logger.Warnf(h.bot.ctx, "Not enough bombs! %s", text)
		return true
	}
	return false
}

func (h *MessageHandler) handleSplitMessage(text string) string {
	if len(text) > 0 && text[0] == '(' {
		pending := h.bot.GetPendingArenaMsg()
		if pending != "" {
			combined := pending + " " + text
			h.logger.Debugf(h.bot.ctx, "Combined split message: %s", combined)
			return combined
		}
	} else {
		h.bot.ClearPendingArenaMsg()
	}
	return text
}

func (h *MessageHandler) checkSplitMessageStart(text, username string) bool {
	if !strings.Contains(text, username+" (") && strings.HasSuffix(text, username) {
		if strings.Contains(text, "Results from the Heist:") ||
			strings.Contains(text, "The dust finally settled") {
			h.bot.SetPendingArenaMsg(text)
			h.logger.Debugf(h.bot.ctx, "Detected split message, buffering...")
			return true
		}
	}
	return false
}

var _ = context.Background
