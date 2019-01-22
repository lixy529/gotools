package logs

import (
	"strings"
)

// formatMsg formate log message.
func formatMsg(n int) string {
	return strings.TrimRight(strings.Repeat("%v"+MsgSep, n), MsgSep)
}

// getLevelName return log level name.
func getLevelName(level int) string {
	switch level {
	case LevelDebug:
		return DebugName
	case LevelInfo:
		return InfoName
	case LevelWarn:
		return WarnName
	case LevelError:
		return ErrorName
	case LevelFatal:
		return FatalName
	default:
		return InfoName
	}
}

// getLevelCode return log level ID.
func getLevelCode(name string) int {
	switch strings.ToUpper(name) {
	case DebugName:
		return LevelDebug
	case InfoName:
		return LevelInfo
	case WarnName:
		return LevelWarn
	case ErrorName:
		return LevelError
	case FatalName:
		return LevelFatal
	default:
		return LevelInfo
	}
}
