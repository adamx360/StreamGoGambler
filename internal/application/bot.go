package application

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"streamgogambler/internal/adapters/logging"
	"streamgogambler/internal/adapters/storage"
	"streamgogambler/internal/domain/gambling"
	"streamgogambler/internal/domain/wallet"
	"streamgogambler/internal/ports"
)

const (
	SlotsJitterFraction     = 0.02
	InitialBombsDelay       = 2 * time.Second
	PostReconnectSlotsDelay = 3 * time.Second
)

var SlotsInterval time.Duration

type BotService struct {
	config ports.ConfigStore
	chat   ports.ChatClient
	wallet *wallet.Wallet
	logger *logging.Logger

	msgHandler *MessageHandler
	cmdHandler *CommandHandler

	mu                 sync.Mutex
	startTime          time.Time
	messagesSent       int
	messagesRecv       int
	reconnectCount     int
	lastReconnect      time.Time
	didGreet           bool
	lastMessageSent    string
	userCmdTimes       map[string]time.Time
	pendingArenaMsg    string
	pendingArenaTime   time.Time
	lastSlotsTime      time.Time
	autoSlotsEnabled   bool
	trustedUsers       map[string]bool
	trustedStore       *storage.TrustedUsersStore
	slotsOffTime       time.Time
	slotsOffCancelChan chan struct{}

	ctx    context.Context
	cancel context.CancelFunc
}

func NewBotService(config ports.ConfigStore, chat ports.ChatClient, logger *logging.Logger, trustedStore *storage.TrustedUsersStore) *BotService {
	trustedUsers, err := trustedStore.Load()
	if err != nil {
		logger.Warnf(context.Background(), "Could not load trusted users: %v, using defaults", err)
		trustedUsers = make(map[string]bool)
	}

	if len(trustedUsers) == 0 {
		if err := trustedStore.Save(trustedUsers); err != nil {
			logger.Warnf(context.Background(), "Could not save default trusted users: %v", err)
		}
	}

	SlotsInterval = time.Duration(config.GetConfig().AutoSlotsInterval) * time.Minute

	return &BotService{
		config:           config,
		chat:             chat,
		wallet:           wallet.New(0),
		logger:           logger,
		userCmdTimes:     make(map[string]time.Time),
		trustedUsers:     trustedUsers,
		trustedStore:     trustedStore,
		autoSlotsEnabled: config.GetConfig().AutoSlotsEnabled,
	}
}

func (s *BotService) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.startTime = time.Now()

	cfg := s.config.GetConfig()

	s.msgHandler = NewMessageHandler(s, s.logger)
	s.cmdHandler = NewCommandHandler(s, s.config, s.logger)

	s.chat.OnConnect(s.onConnect)
	s.chat.OnMessage(s.onMessage)
	s.chat.OnBan(s.onBan)
	s.chat.OnReconnect(s.trackReconnect)
	s.chat.OnNotice(s.onNotice)

	go s.runSlotsLoop()

	go s.runUserCmdTimesCleanup()

	s.chat.Join(cfg.Channel)
	return s.chat.Connect(s.ctx)
}

func (s *BotService) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.chat != nil {
		_ = s.chat.Disconnect()
	}
}

func (s *BotService) onConnect() {
	cfg := s.config.GetConfig()
	s.logger.Infof(s.ctx, "Connected to channel: #%s", cfg.Channel)

	if !s.hasGreeted() || cfg.GreetOnReconnect {
		s.SafeSay(cfg.Channel, cfg.ConnectMessage)
		time.Sleep(InitialBombsDelay)
		s.SafeSay(cfg.Channel, "!bombs")
		s.setGreeted(true)
	}
}

func (s *BotService) onMessage(msg ports.ChatMessage) {
	s.incMessagesRecv()
	s.msgHandler.HandleMessage(msg)
}

func (s *BotService) onBan(event ports.BanEvent) {
	cfg := s.config.GetConfig()
	if event.IsPermanent && cfg.BandOnPerma && cfg.BandMessage != "" {
		s.SafeSay(event.Channel, cfg.BandMessage)
	}
}

func (s *BotService) onNotice(channel, message string) {
	s.logger.Debugf(s.ctx, "NOTICE #%s: %s", channel, message)
	lower := strings.ToLower(message)
	if strings.Contains(lower, "too quick") ||
		strings.Contains(lower, "rate") ||
		strings.Contains(lower, "slow mode") {
		s.retrySend(channel)
	}
}

func (s *BotService) runSlotsLoop() {
	cfg := s.config.GetConfig()
	time.Sleep(2 * InitialBombsDelay)
	s.SafeSay(cfg.Channel, "!slots")
	for {
		d := jitterDuration(SlotsInterval, SlotsJitterFraction)
		t := time.NewTimer(d)
		select {
		case <-s.ctx.Done():
			t.Stop()
			return
		case <-t.C:
		}
		select {
		case <-s.ctx.Done():
			return
		case <-time.After(PostReconnectSlotsDelay):
		}

		if !s.IsAutoSlotsEnabled() {
			continue
		}

		if !s.canPlaySlots() {
			s.logger.Debugf(s.ctx, "Slots cooldown active, skipping this cycle")
			continue
		}

		s.SafeSay(cfg.Channel, "!slots")
	}
}

func (s *BotService) SafeSay(channel, message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}

	cfg := s.config.GetConfig()

	lower := strings.ToLower(message)
	if strings.HasPrefix(lower, "!slots") || strings.HasPrefix(lower, "!heist") || strings.HasPrefix(lower, "!ffa") {
		ok, normalized := s.handleOwnCommands(message, cfg)
		if !ok {
			return
		}
		message = normalized
	}

	if err := s.chat.Say(s.ctx, channel, message); err != nil {
		s.logger.Warnf(s.ctx, "Failed to send message: %v", err)
		return
	}

	s.logger.Infof(s.ctx, "Sent: %s", message)
	s.setLastMessage(message)
	s.incMessagesSent()
}

func (s *BotService) handleOwnCommands(cmd string, cfg ports.BotConfig) (bool, string) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return true, cmd
	}

	base := strings.ToLower(parts[0])
	switch base {
	case "!ffa":
		if s.wallet.Spend(cfg.ArenaCost) {
			s.logger.Infof(s.ctx, "Bot sent !ffa - deducted %d bombs", cfg.ArenaCost)
			return true, "!ffa"
		}
		s.logger.Warnf(s.ctx, "Not enough bombs for !ffa (need %d, have %d)", cfg.ArenaCost, s.wallet.GetBalance())
		return false, cmd

	case "!slots":
		if s.wallet.Spend(cfg.SlotsCost) {
			s.logger.Infof(s.ctx, "Bot sent !slots - deducted %d bombs", cfg.SlotsCost)
			return true, "!slots"
		}
		s.logger.Warnf(s.ctx, "Not enough bombs for !slots (need %d, have %d)", cfg.SlotsCost, s.wallet.GetBalance())
		return false, cmd

	case "!heist":
		amount := cfg.DefaultHeist
		if len(parts) >= 2 {
			if v, err := fmt.Sscanf(parts[1], "%d", &amount); err != nil || v != 1 {
				amount = cfg.DefaultHeist
			}
		}

		amount, _ = gambling.ValidateHeistAmount(amount)
		if amount <= 0 {
			s.logger.Warnf(s.ctx, "Invalid heist amount: %d, skipping", amount)
			return false, cmd
		}

		if s.wallet.Spend(amount) {
			norm := fmt.Sprintf("!heist %d", amount)
			s.logger.Infof(s.ctx, "Bot sent %s - deducted %d bombs", norm, amount)
			return true, norm
		}
		s.logger.Warnf(s.ctx, "Not enough bombs for heist (need %d, have %d)", amount, s.wallet.GetBalance())
		return false, cmd
	}

	return true, cmd
}

func (s *BotService) retrySend(channel string) {
	lastMsg := s.getLastMessage()
	if lastMsg == "" {
		return
	}
	s.logger.Debugf(s.ctx, "Retrying send: %s", lastMsg)
	time.Sleep(1 * time.Second)
	s.SafeSay(channel, lastMsg)
}

func (s *BotService) GetStats() ports.BotStats {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg := s.config.GetConfig()
	uptime := time.Since(s.startTime).Truncate(time.Second)

	return ports.BotStats{
		Status:         "ok",
		Uptime:         uptime.String(),
		UptimeSeconds:  math.Floor(uptime.Seconds()),
		Balance:        s.wallet.GetBalance(),
		MessagesSent:   s.messagesSent,
		MessagesRecv:   s.messagesRecv,
		ReconnectCount: s.reconnectCount,
		Channel:        cfg.Channel,
		Username:       cfg.Username,
	}
}

func (s *BotService) ExecuteCommand(command string) {
	cfg := s.config.GetConfig()

	if s.cmdHandler != nil && strings.HasPrefix(command, cfg.Prefix) {
		cmdPart := strings.TrimPrefix(command, cfg.Prefix)
		parts := strings.Fields(cmdPart)
		if len(parts) > 0 {
			cmdName := strings.ToLower(parts[0])
			if s.cmdHandler.IsInternalCommand(cmdName) {
				s.cmdHandler.HandleCommand(cfg.Username, cfg.Channel, command)
				s.logger.Infof(s.ctx, "Executed internal command: %s", command)
				return
			}
		}
	}

	s.SafeSay(cfg.Channel, command)
}

func (s *BotService) Wallet() *wallet.Wallet {
	return s.wallet
}

func (s *BotService) Config() ports.ConfigStore {
	return s.config
}

func (s *BotService) Logger() *logging.Logger {
	return s.logger
}

func (s *BotService) hasGreeted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.didGreet
}

func (s *BotService) setGreeted(v bool) {
	s.mu.Lock()
	s.didGreet = v
	s.mu.Unlock()
}

func (s *BotService) getLastMessage() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastMessageSent
}

func (s *BotService) setLastMessage(msg string) {
	s.mu.Lock()
	s.lastMessageSent = msg
	s.mu.Unlock()
}

func (s *BotService) incMessagesSent() {
	s.mu.Lock()
	s.messagesSent++
	s.mu.Unlock()
}

func (s *BotService) incMessagesRecv() {
	s.mu.Lock()
	s.messagesRecv++
	s.mu.Unlock()
}

func (s *BotService) trackReconnect() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	s.reconnectCount++

	if !s.lastReconnect.IsZero() && now.Sub(s.lastReconnect) > 10*time.Minute {
		s.reconnectCount = 1
	}

	s.lastReconnect = now
	return s.reconnectCount > 5
}

func (s *BotService) IsUserRateLimited(username string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	const userCommandRateLimit = 2 * time.Second
	now := time.Now()
	lowerName := strings.ToLower(username)

	if lastTime, exists := s.userCmdTimes[lowerName]; exists {
		if now.Sub(lastTime) < userCommandRateLimit {
			return true
		}
	}

	s.userCmdTimes[lowerName] = now
	return false
}

func (s *BotService) runUserCmdTimesCleanup() {
	const cleanupInterval = 10 * time.Minute
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.cleanupUserCmdTimes()
		}
	}
}

func (s *BotService) cleanupUserCmdTimes() {
	s.mu.Lock()
	defer s.mu.Unlock()

	const maxAge = 5 * time.Minute
	now := time.Now()

	for user, lastTime := range s.userCmdTimes {
		if now.Sub(lastTime) > maxAge {
			delete(s.userCmdTimes, user)
		}
	}
}

func (s *BotService) SetPendingArenaMsg(msg string) {
	s.mu.Lock()
	s.pendingArenaMsg = msg
	s.pendingArenaTime = time.Now()
	s.mu.Unlock()
}

func (s *BotService) GetPendingArenaMsg() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	const pendingTimeout = 5 * time.Second

	if s.pendingArenaMsg == "" {
		return ""
	}

	if time.Since(s.pendingArenaTime) > pendingTimeout {
		s.pendingArenaMsg = ""
		return ""
	}

	msg := s.pendingArenaMsg
	s.pendingArenaMsg = ""
	return msg
}

func (s *BotService) ClearPendingArenaMsg() {
	s.mu.Lock()
	s.pendingArenaMsg = ""
	s.mu.Unlock()
}

func (s *BotService) canPlaySlots() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lastSlotsTime.IsZero() {
		return true
	}
	return time.Since(s.lastSlotsTime) >= SlotsInterval
}

func (s *BotService) RecordSlotsPlayed() {
	s.mu.Lock()
	s.lastSlotsTime = time.Now()
	s.mu.Unlock()
}

func (s *BotService) IsAutoSlotsEnabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.autoSlotsEnabled
}

func (s *BotService) SetAutoSlots(enabled bool) {
	s.mu.Lock()
	s.autoSlotsEnabled = enabled
	s.mu.Unlock()
}

func (s *BotService) IsUserTrusted(username string) bool {
	cfg := s.config.GetConfig()
	if strings.EqualFold(username, cfg.Username) {
		return true
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trustedUsers[strings.ToLower(username)]
}

func (s *BotService) AddTrustedUser(username string) {
	s.mu.Lock()
	s.trustedUsers[strings.ToLower(username)] = true
	usersCopy := make(map[string]bool)
	for k, v := range s.trustedUsers {
		usersCopy[k] = v
	}
	s.mu.Unlock()

	if err := s.trustedStore.Save(usersCopy); err != nil {
		s.logger.Warnf(s.ctx, "Could not save trusted users: %v", err)
	}
}

func (s *BotService) RemoveTrustedUser(username string) {
	s.mu.Lock()
	delete(s.trustedUsers, strings.ToLower(username))
	usersCopy := make(map[string]bool)
	for k, v := range s.trustedUsers {
		usersCopy[k] = v
	}
	s.mu.Unlock()

	if err := s.trustedStore.Save(usersCopy); err != nil {
		s.logger.Warnf(s.ctx, "Could not save trusted users: %v", err)
	}
}

func (s *BotService) GetTrustedUsers() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	users := make([]string, 0, len(s.trustedUsers))
	for user := range s.trustedUsers {
		users = append(users, user)
	}
	return users
}

func (s *BotService) ScheduleSlotsOff(offTime time.Time) {
	s.mu.Lock()

	if s.slotsOffCancelChan != nil {
		close(s.slotsOffCancelChan)
	}

	s.slotsOffTime = offTime
	s.slotsOffCancelChan = make(chan struct{})
	cancelChan := s.slotsOffCancelChan

	s.mu.Unlock()

	go func() {
		duration := time.Until(offTime)
		if duration <= 0 {
			return
		}

		timer := time.NewTimer(duration)
		defer timer.Stop()

		select {
		case <-timer.C:
			s.SetAutoSlots(false)
			s.logger.Infof(s.ctx, "Auto slots turned off (scheduled)")
			s.mu.Lock()
			s.slotsOffTime = time.Time{}
			s.slotsOffCancelChan = nil
			s.mu.Unlock()
		case <-cancelChan:
			// Canceled
		case <-s.ctx.Done():
			// Bot shutting down
		}
	}()
}

func (s *BotService) CancelSlotsOffSchedule() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.slotsOffCancelChan == nil {
		return false
	}

	close(s.slotsOffCancelChan)
	s.slotsOffCancelChan = nil
	s.slotsOffTime = time.Time{}
	return true
}

func (s *BotService) GetSlotsOffTime() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.slotsOffTime
}

func jitterDuration(base time.Duration, fraction float64) time.Duration {
	jitter := float64(base) * fraction * (2*rand.Float64() - 1)
	return base + time.Duration(math.Abs(jitter))
}
