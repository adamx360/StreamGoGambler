package ports

type BotStats struct {
	Status         string  `json:"status"`
	Uptime         string  `json:"uptime"`
	UptimeSeconds  float64 `json:"uptime_seconds"`
	Balance        int     `json:"bombs"`
	MessagesSent   int     `json:"messages_sent"`
	MessagesRecv   int     `json:"messages_received"`
	ReconnectCount int     `json:"reconnect_count"`
	Channel        string  `json:"channel"`
	Username       string  `json:"username"`
}

type StatsProvider interface {
	GetStats() BotStats
}
