package conf

const (
	LenStackBuf   = 4096
	ConsolePrompt = "Leaf# "
)

var (
	// log
	LogLevel string
	LogPath  string
	LogFlag  int

	// console
	ConsolePort int
	ProfilePath string

	// cluster
	ListenAddr      string
	ConnAddrs       []string
	PendingWriteNum int
)
