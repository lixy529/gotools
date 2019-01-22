package logs

// Log level
const (
	LevelDebug = 1
	LevelInfo  = 2
	LevelWarn  = 3
	LevelError = 4
	LevelFatal = 5
)

const (
	DebugName = "DEBUG"
	InfoName  = "INFO"
	WarnName  = "WARN"
	ErrorName = "ERROR"
	FatalName = "FATAL"
)

// Log output type
const (
	AdapterFrame    = "frame" // Write file, use for frameworks
	AdapterConsole  = "console"
	AdapterFile     = "file"
	AdapterSyslogNg = "syslog"
)

const (
	MsgSep   = "\t"
	DefDepth = 3
)

// LoggerInter
type Logger interface {
	Init(config string) error
	Destroy()
	Flush()
	WriteMsg(level int, fmtStr string, v ...interface{}) error

	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})

	Debugf(fmtStr string, v ...interface{})
	Infof(fmtStr string, v ...interface{})
	Warnf(fmtStr string, v ...interface{})
	Errorf(fmtStr string, v ...interface{})
	Fatalf(fmtStr string, v ...interface{})
}

var adapters = make(map[string]Logger)

// Register Register an adapter.
func Register(name string, log Logger) {
	if log == nil {
		panic("logs: Register provide is nil")
	}
	if _, dup := adapters[name]; dup {
		panic("logs: Register called twice for provider " + name)
	}
	adapters[name] = log
}

// Log return a Logger.
func Log(name string) Logger {
	logger, ok := adapters[name]
	if ok {
		return logger
	}

	return nil
}

// GetLevelNameById get level name from log level ID.
func GetLevelNameById(levelId int) string {
	if levelId == LevelDebug {
		return DebugName
	} else if levelId == LevelInfo {
		return InfoName
	} else if levelId == LevelWarn {
		return WarnName
	}  else if levelId == LevelError {
		return ErrorName
	}  else if levelId == LevelFatal {
		return FatalName
	}

	return ""
}

// GetLevelIdByName get level ID from log level name.
func GetLevelIdByName(levelName string) int {
	if levelName == DebugName {
		return LevelDebug
	} else if levelName == InfoName {
		return LevelInfo
	} else if levelName == WarnName {
		return LevelWarn
	}  else if levelName == ErrorName {
		return LevelError
	}  else if levelName == FatalName {
		return LevelFatal
	}

	return -1
}
