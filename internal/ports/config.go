package ports

type BotConfig struct {
	Username string
	Channel  string

	Prefix        string
	StatusCommand string

	ConnectMessage   string
	BandMessage      string
	BandOnPerma      bool
	GreetOnReconnect bool

	BossBotName   string
	AutoResponses map[string]string

	DefaultHeist int
	SlotsCost    int
	ArenaCost    int

	AutoSlotsEnabled  bool
	AutoSlotsInterval int

	PointsAsDelta bool

	SayBucketSize int
	SayRefillMs   int

	LogLevel   string
	HealthPort int

	GUIEnabled   bool
	MaxLogsLines int
}

type ConfigReader interface {
	GetConfig() BotConfig

	GetOAuth() string
}

type ConfigWriter interface {
	UpdateHeist(amount int) error
}

type ConfigStore interface {
	ConfigReader
	ConfigWriter
}
