// Print log
package logs

import (
	"github.com/lixy529/gotools/utils"
	"encoding/json"
	"fmt"
)

type ConsoleLogs struct {
	Level    int  `json:"level"`    // Log level
	ShowCall bool `json:"showcall"` // Display the file name and line number of calling code
	Depth    int  `json:"depth"`    // Call function depth
}

// Init initialization configuration.
// Eg:
// {
// "level":1,
// "showcall":true,
// "depth":3
// }
func (l *ConsoleLogs) Init(config string) error {
	l.Level = LevelDebug
	l.ShowCall = false
	l.Depth = DefDepth

	if len(config) == 0 {
		return nil
	}

	err := json.Unmarshal([]byte(config), l)
	if err != nil {
		return err
	}

	return nil
}

// WriteMsg write log message.
// Eg: WriteMsg(LevelInfo, "%s-%s", "aa", "bb)
func (l *ConsoleLogs) WriteMsg(level int, fmtStr string, v ...interface{}) error {
	if level < l.Level {
		return nil
	}

	msg := fmt.Sprintf(fmtStr, v...)
	curTime := utils.CurTime()
	levelName := getLevelName(level)

	strTmpFmt := "%s" + MsgSep + "[%s]" + MsgSep

	if l.ShowCall {
		file, line := utils.GetCall(l.Depth)
		if level == LevelWarn {
			fmt.Printf(strTmpFmt+"%c[1;00;33m%s%c[0m"+MsgSep+"(%s:%d)\n", curTime, levelName, 0x1B, msg, 0x1B, file, line)
		} else if level > LevelWarn {
			fmt.Printf(strTmpFmt+"%c[1;00;31m%s%c[0m"+MsgSep+"(%s:%d)\n", curTime, levelName, 0x1B, msg, 0x1B, file, line)
		} else {
			fmt.Printf(strTmpFmt+"%s"+MsgSep+"(%s:%d)\n", curTime, levelName, msg, file, line)
		}
	} else {
		if level == LevelWarn {
			fmt.Printf(strTmpFmt+"%c[1;00;33m%s%c[0m\n", curTime, levelName, 0x1B, msg, 0x1B)
		} else if level > LevelWarn {
			fmt.Printf(strTmpFmt+"%c[1;00;31m%s%c[0m\n", curTime, levelName, 0x1B, msg, 0x1B)
		} else {
			fmt.Printf(strTmpFmt+"%s\n", curTime, levelName, msg)
		}
	}

	return nil
}

// Debug write debug log.
func (l *ConsoleLogs) Debug(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelDebug, fmtStr, v...)
}

// Info write info log.
func (l *ConsoleLogs) Info(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelInfo, fmtStr, v...)
}

// Warn write warn log.
func (l *ConsoleLogs) Warn(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelWarn, fmtStr, v...)
}

// Error write error log.
func (l *ConsoleLogs) Error(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelError, fmtStr, v...)
}

// Fatal write fatal log.
func (l *ConsoleLogs) Fatal(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelFatal, fmtStr, v...)
}

// Debugf write debug log.
func (l *ConsoleLogs) Debugf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelDebug, fmtStr, v...)
}

// Infof write info log.
func (l *ConsoleLogs) Infof(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelInfo, fmtStr, v...)
}

// Warnf write warn log.
func (l *ConsoleLogs) Warnf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelWarn, fmtStr, v...)
}

// Errorf write error log.
func (l *ConsoleLogs) Errorf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelError, fmtStr, v...)
}

// Fatalf write fatal log.
func (l *ConsoleLogs) Fatalf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelFatal, fmtStr, v...)
}

// Destroy
func (l *ConsoleLogs) Destroy() {
}

// Flush
func (l *ConsoleLogs) Flush() {
}

// init register adapter.
func init() {
	Register(AdapterConsole, &ConsoleLogs{Level: LevelDebug})
}
