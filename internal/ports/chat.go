package ports

import "context"

type ChatMessage struct {
	ID       string
	UserName string
	UserID   string
	Channel  string
	Text     string
}

type BanEvent struct {
	Channel     string
	UserName    string
	Duration    int // 0 for permanent ban
	IsPermanent bool
}

type MessageSender interface {
	Say(ctx context.Context, channel, message string) error
}

type ChatClient interface {
	MessageSender

	Connect(ctx context.Context) error

	Disconnect() error

	Join(channel string)

	OnMessage(handler func(ChatMessage))

	OnConnect(handler func())

	OnBan(handler func(BanEvent))

	OnReconnect(handler func() bool)

	OnNotice(handler func(channel, message string))
}
