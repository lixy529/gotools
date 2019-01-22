// Log write to file
package logs

import (
	"github.com/lixy529/gotools/utils"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	SizeUnit    = 1024 * 1024
	DefMaxLines = 1000000
	DefMaxSize  = 256 //256 MB *SizeUnit
	DefPerm     = "0660"
)

// logsFile
type FileLogs struct {
	sync.RWMutex

	FilePath   string `json:"filepath"`
	FileName   string `json:"filename"`
	fullFile   string // Log file full path, sub directory[YYYYMMDD] will be created in FilePath.
	fileWriter *os.File

	MaxLines int    `json:"maxlines"` // Maximum number of lines, 0 unlimited.
	curLines int                      // Current number of lines.
	MaxSize  int    `json:"maxsize"`  // Maximum file size, unit M, 0 unlimited.
	curSize  int                      // Current file size.
	Perm     string `json:"perm"`     // File permissions.
	openTime time.Time                // Open file time.

	Level    int  `json:"level"`
	ShowCall bool `json:"showcall"`
	Depth    int  `json:"depth"`
}

// newFileWriter create a FileLogWriter returning as LoggerInterface.
func newLogsFile() Logger {
	l := &FileLogs{
		MaxLines: DefMaxLines,
		MaxSize:  DefMaxSize * SizeUnit,
		Perm:     DefPerm,
		Level:    LevelDebug,
		ShowCall: false,
		Depth:    DefDepth,
	}
	return l
}

// Init initialization configuration.
// Eg:
// {
// "filepath":"/tmp/log"
// "filename":"web.log",
// "maxLines":10000,
// "maxsize":500,
// "perm":"0600"
// }
func (l *FileLogs) Init(config string) error {
	l.MaxLines = DefMaxLines
	l.MaxSize = DefMaxSize
	l.Perm = DefPerm
	l.Level = LevelDebug
	l.ShowCall = false
	l.Depth = DefDepth

	if len(config) == 0 {
		l.MaxSize = l.MaxSize * SizeUnit
		return nil
	}

	err := json.Unmarshal([]byte(config), l)
	if err != nil {
		return err
	}

	if l.FilePath == "" && l.FileName == "" {
		return errors.New("FileLogs: filepath or filename is empty")
	}

	l.MaxSize = l.MaxSize * SizeUnit

	return l.startLogger()
}

// startLogger start a new log file.
func (l *FileLogs) startLogger() error {
	f, err := l.createLogFile()
	if err != nil {
		return err
	}

	if l.fileWriter != nil {
		l.fileWriter.Close()
	}

	l.fileWriter = f

	return l.initFd()
}

// initFd initialize log file.
func (l *FileLogs) initFd() error {
	fd := l.fileWriter
	fInfo, err := fd.Stat()
	if err != nil {
		return fmt.Errorf("FileLogs: Get stat err: %s\n", err)
	}
	l.curSize = int(fInfo.Size())
	l.openTime = time.Now()
	if fInfo.Size() > 0 {
		count, err := l.lines()
		if err != nil {
			return err
		}
		l.curLines = count
	} else {
		l.curLines = 0
	}

	// Update logs file at 0:00 am every day.
	//go l.dailyChange(l.openTime)

	return nil
}

// dailyChange update logs file at 0:00 am every day.
func (l *FileLogs) dailyChange(openTime time.Time) {
	y, m, d := openTime.Add(24 * time.Hour).Date()
	nextDay := time.Date(y, m, d, 0, 0, 0, 0, openTime.Location())
	tm := time.NewTimer(time.Duration(nextDay.UnixNano() - openTime.UnixNano() + 100))
	select {
	case <-tm.C:
		l.Lock()
		if l.needChange(time.Now().Day()) {
			if err := l.doChange(); err != nil {
				fmt.Fprintf(os.Stderr, "FileLogs: %q[%s]\n", l.fullFile, err)
			}
		}
		l.Unlock()
	}
}

// lines returns number of lines.
func (l *FileLogs) lines() (int, error) {
	fd, err := os.Open(l.fullFile)
	if err != nil {
		return 0, err
	}
	defer fd.Close()

	buf := make([]byte, 32768) // 32k
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := fd.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}

		count += bytes.Count(buf[:c], lineSep)

		if err == io.EOF {
			break
		}
	}

	return count, nil
}

// needChange need update log file.
func (l *FileLogs) needChange(day int) bool {
	return (l.MaxLines > 0 && l.curLines >= l.MaxLines) ||
		(l.MaxSize > 0 && l.curSize >= l.MaxSize) ||
		(day != l.openTime.Day())
}

// backFile backup log file.
func (l *FileLogs) backFile() error {
	var err error

	_, err = os.Lstat(l.fullFile)
	if err != nil {
		return nil
	}

	num := 1
	fName := ""
	for ; err == nil && num <= 9999; num++ {
		fName = fmt.Sprintf("%s_%04d", l.fullFile, num)
		_, err = os.Lstat(fName)
		if err != nil {
			goto RENAME
		}
	}

	return fmt.Errorf("FileLogs: Cannot find free log number to rename %s\n", l.fullFile)

RENAME:
	l.fileWriter.Close()
	err = os.Rename(l.fullFile, fName)

	return err
}

// doChange update new log file, log write to the new file.
func (l *FileLogs) doChange() error {
	err := l.backFile()
	if err != nil {
		return err
	}

	err = l.startLogger()
	if err != nil {
		return fmt.Errorf("FileLogs: startLogger error %s\n", err)
	}

	return err
}

// createLogFile create a new log file.
func (l *FileLogs) createLogFile() (*os.File, error) {
	perm, err := strconv.ParseInt(l.Perm, 8, 64)
	if err != nil {
		return nil, err
	}

	// 拼文件路径
	l.fullFile = path.Join(l.FilePath, utils.CurTime("060102"), l.FileName)
	err = utils.MkDir(l.fullFile, 0760, true)
	if err != nil {
		return nil, err
	}

	fd, err := os.OpenFile(l.fullFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(perm))
	if err == nil {
		os.Chmod(l.fullFile, os.FileMode(perm))
	}

	return fd, err
}

// WriteMsg write log message.
// Eg: WriteMsg(LevelInfo, "%s-%s", "aa", "bb)
func (l *FileLogs) WriteMsg(level int, fmtStr string, v ...interface{}) error {
	if level < l.Level {
		return nil
	}

	now := time.Now()
	curTime := now.Format("2006-01-02 15:04:05")
	strMsg := fmt.Sprintf(fmtStr, v...)
	levelName := getLevelName(level)
	var arrMsg []string
	arrMsg = append(arrMsg, curTime)
	arrMsg = append(arrMsg, "["+levelName+"]")
	arrMsg = append(arrMsg, strMsg)
	if l.ShowCall {
		file, line := utils.GetCall(l.Depth)
		arrMsg = append(arrMsg, fmt.Sprintf("(%s:%d)", file, line))
	}
	msg := strings.Join(arrMsg, MsgSep)
	msg = msg + "\n"

	l.RLock()
	if l.needChange(now.Day()) {
		l.RUnlock()
		l.Lock()
		if l.needChange(now.Day()) {
			if err := l.doChange(); err != nil {
				fmt.Fprintf(os.Stderr, "FileLogs: %q[%s]\n", l.fullFile, err)
			}
		}
		l.Unlock()
	} else {
		l.RUnlock()
	}

	l.Lock()
	_, err := l.fileWriter.Write([]byte(msg))
	if err == nil {
		l.curLines++
		l.curSize += len(msg)
	}
	l.Unlock()

	return nil
}

// Debug write debug log.
func (l *FileLogs) Debug(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelDebug, fmtStr, v...)
}

// Info write info log.
func (l *FileLogs) Info(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelInfo, fmtStr, v...)
}

// Warn write warn log.
func (l *FileLogs) Warn(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelWarn, fmtStr, v...)
}

// Error write error log.
func (l *FileLogs) Error(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelError, fmtStr, v...)
}

// Fatal write fatal log.
func (l *FileLogs) Fatal(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelFatal, fmtStr, v...)
}

// Debugf write debug log.
func (l *FileLogs) Debugf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelDebug, fmtStr, v...)
}

// Infof write info log.
func (l *FileLogs) Infof(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelInfo, fmtStr, v...)
}

// Warnf write warn log.
func (l *FileLogs) Warnf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelWarn, fmtStr, v...)
}

// Errorf write error log.
func (l *FileLogs) Errorf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelError, fmtStr, v...)
}

// Fatalf write fatal log.
func (l *FileLogs) Fatalf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelFatal, fmtStr, v...)
}

// Destroy close file.
func (l *FileLogs) Destroy() {
	l.fileWriter.Close()
	l.fileWriter = nil
}

// Flush flush file.
func (l *FileLogs) Flush() {
	l.fileWriter.Sync()
}

// init register adapter.
func init() {
	Register(AdapterFile, newLogsFile())
	Register(AdapterFrame, newLogsFile())
}
