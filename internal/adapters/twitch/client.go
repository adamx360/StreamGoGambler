package twitch

import (
	"context"
	"errors"
	"time"

	"github.com/gempir/go-twitch-irc/v4"

	"streamgogambler/internal/adapters/logging"
	"streamgogambler/internal/ports"
)

const (
	InitialRetryDelay = 1 * time.Second
	MaxRetryDelay     = 5 * time.Minute
	RetryMultiplier   = 2.0
	MaxRetryAttempts  = 0 // 0 = unlimited

	DefaultBucketSize = 20
	DefaultRefillMs   = 150
	SafeSayTimeout    = 30 * time.Second
)

var (
	ErrSayTimeout   = errors.New("say timeout waiting for token")
	ErrDisconnected = errors.New("client disconnected")
)

type Client struct {
	irc    *twitch.Client
	logger *logging.Logger

	tokens     chan struct{}
	bucketSize int
	refillMs   int

	onMessage   func(ports.ChatMessage)
	onConnect   func()
	onBan       func(ports.BanEvent)
	onReconnect func() bool
	onNotice    func(channel, message string)

	ctx    context.Context
	cancel context.CancelFunc
}

type ClientOption func(*Client)

func WithBucketSize(size int) ClientOption {
	return func(c *Client) {
		if size > 0 {
			c.bucketSize = size
		}
	}
}

func WithRefillMs(ms int) ClientOption {
	return func(c *Client) {
		if ms > 0 {
			c.refillMs = ms
		}
	}
}

func WithLogger(logger *logging.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

func NewClient(username, oauth string, opts ...ClientOption) *Client {
	c := &Client{
		irc:        twitch.NewClient(username, "oauth:"+oauth),
		bucketSize: DefaultBucketSize,
		refillMs:   DefaultRefillMs,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.logger == nil {
		c.logger = logging.New(logging.LevelInfo)
	}

	c.tokens = make(chan struct{}, c.bucketSize)
	c.setupHandlers()

	return c
}

func (c *Client) setupHandlers() {
	c.irc.OnConnect(func() {
		if c.onConnect != nil {
			c.onConnect()
		}
	})

	c.irc.OnPrivateMessage(func(msg twitch.PrivateMessage) {
		if c.onMessage != nil {
			c.onMessage(ports.ChatMessage{
				ID:       msg.ID,
				UserName: msg.User.Name,
				UserID:   msg.User.ID,
				Channel:  msg.Channel,
				Text:     msg.Message,
			})
		}
	})

	c.irc.OnClearChatMessage(func(msg twitch.ClearChatMessage) {
		if c.onBan != nil {
			c.onBan(ports.BanEvent{
				Channel:     msg.Channel,
				UserName:    msg.TargetUsername,
				Duration:    msg.BanDuration,
				IsPermanent: msg.BanDuration == 0,
			})
		}
	})

	c.irc.OnReconnectMessage(func(_ twitch.ReconnectMessage) {
		if c.onReconnect != nil {
			if c.onReconnect() {
				c.logger.Errorf(c.ctx, "High reconnect frequency - check network connection!")
			}
		}
		c.logger.Warnf(c.ctx, "Twitch requested reconnect - reconnecting...")
	})

	c.irc.OnNoticeMessage(func(msg twitch.NoticeMessage) {
		if c.onNotice != nil {
			c.onNotice(msg.Channel, msg.Message)
		}
	})
}

func (c *Client) Connect(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)

	go c.runTokenRefiller()

	delay := InitialRetryDelay
	attempt := 0

	for {
		attempt++
		err := c.irc.Connect()
		if err == nil {
			return nil
		}

		select {
		case <-c.ctx.Done():
			c.logger.Infof(c.ctx, "Connection interrupted during shutdown")
			return c.ctx.Err()
		default:
		}

		if MaxRetryAttempts > 0 && attempt >= MaxRetryAttempts {
			c.logger.Errorf(c.ctx, "Max connection attempts (%d) exceeded: %v", MaxRetryAttempts, err)
			return err
		}

		c.logger.Warnf(c.ctx, "Connection error (attempt %d): %v. Retrying in %v...", attempt, err, delay)

		select {
		case <-c.ctx.Done():
			c.logger.Infof(c.ctx, "Connection interrupted during shutdown")
			return c.ctx.Err()
		case <-time.After(delay):
		}

		delay = time.Duration(float64(delay) * RetryMultiplier)
		if delay > MaxRetryDelay {
			delay = MaxRetryDelay
		}
	}
}

func (c *Client) runTokenRefiller() {
	ticker := time.NewTicker(time.Duration(c.refillMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			select {
			case c.tokens <- struct{}{}:
			default:
			}
		}
	}
}

func (c *Client) Disconnect() error {
	if c.cancel != nil {
		c.cancel()
	}
	return c.irc.Disconnect()
}

func (c *Client) Join(channel string) {
	c.irc.Join(channel)
}

func (c *Client) Say(ctx context.Context, channel, message string) error {
	timeout := time.NewTimer(SafeSayTimeout)
	defer timeout.Stop()

	select {
	case <-c.tokens:
	case <-ctx.Done():
		return ctx.Err()
	case <-timeout.C:
		c.logger.Warnf(ctx, "Say timeout after %v for: %s", SafeSayTimeout, message)
		return ErrSayTimeout
	}

	c.irc.Say(channel, message)
	return nil
}

func (c *Client) OnMessage(handler func(ports.ChatMessage)) {
	c.onMessage = handler
}

func (c *Client) OnConnect(handler func()) {
	c.onConnect = handler
}

func (c *Client) OnBan(handler func(ports.BanEvent)) {
	c.onBan = handler
}

func (c *Client) OnReconnect(handler func() bool) {
	c.onReconnect = handler
}

func (c *Client) OnNotice(handler func(channel, message string)) {
	c.onNotice = handler
}
