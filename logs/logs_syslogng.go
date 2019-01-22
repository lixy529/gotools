// Log write to syslog-ng
package logs

import (
	"github.com/lixy529/gotools/utils"
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
	"strings"
)

const (
	DefSockFile    = "/var/run/syslog/syslog-ng.sock"
	DefMaxConns    = 0
	DefMaxIdle     = 100
	DefIdleTimeout = 3
	DefConnTimeout = 50
	DefReqTimeout  = 50
)

// SyslogNgLogs
type SyslogNgLogs struct {
	Addr        string        `json:"addr"`        // syslog address, sock file or ip and port(ip:port)
	addrType    string                             // address type, tcp or unix
	LocalFile   string        `json:"localfile"`   // Write local files when an exception occurs
	MaxConns    int           `json:"maxconns"`    // Maximum number of connections, default is 0, 0 is unlimited, if less than 0, don't use pool.
	MaxIdle     int           `json:"maxidle"`     // Maximum number of idle connections, default is 100, 0 is unlimited.
	IdleTimeout time.Duration `json:"idletimeout"` // Idle connection timeout time, default 3 seconds, 0 unlimited, unit is seconds.
	Wait        bool          `json:"wait"`        // Wait or not when hasn't idle connections and the maximum number of connections is reached, default is no waiting and return a error.
	Level       int           `json:"level"`
	ShowCall    bool          `json:"showcall"`
	Depth       int           `json:"depth"`

	pool *SockPool
	mux  sync.Mutex
}

// Init initialization configuration.
// Eg:
// {
// "addr":"/var/run/php-syslog-ng.sock",
// "localfile":"/tmp/gomessages",
// "maxconns":20
// "maxidle":10
// "idletimeout":3
// "level":1,
// "showcall":true,
// "depth":3
// }
func (l *SyslogNgLogs) Init(config string) error {
	l.Addr = DefSockFile
	l.MaxConns = DefMaxConns
	l.MaxIdle = DefMaxIdle
	l.IdleTimeout = DefIdleTimeout
	l.Wait = false
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

	if l.Addr == "" {
		l.Addr = DefSockFile
	}

	// check address, sock file or ip:port
	if strings.Contains(l.Addr, ":") {
		l.addrType = "tcp"
	} else {
		l.addrType = "unix"
		isFile, err := utils.IsFile(l.Addr)
		if err != nil {
			return err
		} else if !isFile {
			return fmt.Errorf("SyslogNgLogs: [%s] is not file.", l.Addr)
		}
	}

	// init connect pool
	if l.MaxConns >= 0 {
		l.pool = NewSockPool(l.Addr, l.addrType, l.MaxConns, l.MaxIdle, l.IdleTimeout, l.Wait)
		if l.pool == nil {
			return fmt.Errorf("SyslogNgLogs: NewSockPool failed [%s].", l.Addr)
		}
	}

	return nil
}

// WriteMsg write log message.
// Eg: WriteMsg(LevelInfo, "%s-%s", "aa", "bb)
func (l *SyslogNgLogs) WriteMsg(level int, fmtStr string, v ...interface{}) error {
	if level < l.Level {
		return nil
	}

	msg := fmt.Sprintf(fmtStr, v...)
	if l.ShowCall {
		file, line := utils.GetCall(l.Depth)
		msg += MsgSep + fmt.Sprintf("(%s:%d)", file, line)
	}
	msg += "\n"

	var conn net.Conn
	var err error
	forceClose := false

	if l.pool == nil {
		// 不使用连接池
		conn, err = net.DialTimeout(l.addrType, l.Addr, DefConnTimeout*time.Millisecond)
		if err != nil || conn == nil {
			l.writeLocalFile(msg)
			return err
		} else {
			conn.SetDeadline(time.Now().Add(DefReqTimeout * time.Millisecond))
		}
		defer conn.Close()
	} else {
		// 使用连接池
		conn, err = l.pool.Get()
		defer l.pool.Put(conn, forceClose)
		if err != nil {
			l.writeLocalFile(msg)
			forceClose = true
			return err
		}
	}

	// write log
	_, err = conn.Write([]byte(msg))
	if err != nil {
		l.writeLocalFile(msg)
		forceClose = true
		return err
	}

	return nil
}

// writeLocalFile write log to local file.
func (l *SyslogNgLogs) writeLocalFile(msg string) error {
	l.mux.Lock()
	defer l.mux.Unlock()

	fd, err := os.OpenFile(l.LocalFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.FileMode(0660))
	if err != nil {
		return err
	}

	_, err = fd.Write([]byte(msg))
	fd.Close()
	return err
}

// Debug write debug log.
func (l *SyslogNgLogs) Debug(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelDebug, fmtStr, v...)
}

// Info write info log.
func (l *SyslogNgLogs) Info(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelInfo, fmtStr, v...)
}

// Warn write warn log.
func (l *SyslogNgLogs) Warn(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelWarn, fmtStr, v...)
}

// Error write error log.
func (l *SyslogNgLogs) Error(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelError, fmtStr, v...)
}

// Fatal write fatal log.
func (l *SyslogNgLogs) Fatal(v ...interface{}) {
	fmtStr := formatMsg(len(v))
	l.WriteMsg(LevelFatal, fmtStr, v...)
}

// Debugf write debug log.
func (l *SyslogNgLogs) Debugf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelDebug, fmtStr, v...)
}

// Infof write info log.
func (l *SyslogNgLogs) Infof(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelInfo, fmtStr, v...)
}

// Warnf write warn log.
func (l *SyslogNgLogs) Warnf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelWarn, fmtStr, v...)
}

// Errorf write error log.
func (l *SyslogNgLogs) Errorf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelError, fmtStr, v...)
}

// Fatalf write fatal log.
func (l *SyslogNgLogs) Fatalf(fmtStr string, v ...interface{}) {
	l.WriteMsg(LevelFatal, fmtStr, v...)
}

// Destroy close connect pool.
func (l *SyslogNgLogs) Destroy() {
	if l.pool != nil {
		l.pool.Close()
	}
}

// Flush
func (l *SyslogNgLogs) Flush() {
}

// init register adapter.
func init() {
	Register(AdapterSyslogNg, &SyslogNgLogs{Level: LevelDebug})
}

var nowFunc = time.Now
var ErrPoolExhausted = errors.New("SockPool: Connection pool exhausted")

// IdleConn idle connection struct.
type IdleConn struct {
	c net.Conn
	t time.Time
}

// SockPool connection pool.
type SockPool struct {
	addr        string
	addrType    string
	maxConns    int
	maxIdle     int
	idleTimeout time.Duration

	idleConns list.List

	mu       sync.Mutex
	cond     *sync.Cond
	wait     bool
	curConns int
	closed   bool
}

// NewSockPool new a connection pool.
func NewSockPool(addr, addrType string, maxConns, maxIdle int, idleTimeout time.Duration, wait bool) *SockPool {
	return &SockPool{
		addr:        addr,
		addrType:    addrType,
		maxConns:    maxConns,
		maxIdle:     maxIdle,
		idleTimeout: idleTimeout * time.Second,
		curConns:    0,
		closed:      false,
		wait:        wait,
	}
}

// Get returns a connection.
func (p *SockPool) Get() (net.Conn, error) {
	p.mu.Lock()

	// check timeout connection
	if timeout := p.idleTimeout; timeout > 0 {
		for i, n := 0, p.idleConns.Len(); i < n; i++ {
			e := p.idleConns.Back()
			if e == nil {
				break
			}
			ic := e.Value.(IdleConn)
			if ic.t.Add(timeout).After(nowFunc()) {
				break
			}
			p.idleConns.Remove(e)
			p.release()
			p.mu.Unlock()
			ic.c.Close()
			p.mu.Lock()
		}
	}

	for {
		// Take an idle connection if there is idle.
		for i, n := 0, p.idleConns.Len(); i < n; i++ {
			e := p.idleConns.Front()
			if e == nil {
				break
			}
			ic := e.Value.(IdleConn)
			p.idleConns.Remove(e)
			// todo: Test the availability of connections.
			p.mu.Unlock()
			return ic.c, nil
		}

		// Check that the connection pool is closed.
		if p.closed {
			p.mu.Unlock()
			return nil, errors.New("SockPool: Pool is closed")
		}

		// Create a new connection, if there is no idle connection and the number of connections is less than the maximum number.
		if p.maxConns == 0 || p.curConns < p.maxConns {
			p.curConns += 1
			p.mu.Unlock()
			c, err := net.DialTimeout(p.addrType, p.addr, DefConnTimeout*time.Millisecond)
			if err != nil || c == nil {
				p.mu.Lock()
				p.release()
				p.mu.Unlock()
				c = nil
			} else {
				c.SetDeadline(time.Now().Add(DefReqTimeout * time.Millisecond))
			}

			return c, err
		}

		// If you don't wait to return the error directly
		if !p.wait {
			p.mu.Unlock()
			return nil, ErrPoolExhausted
		}

		// Waiting for connection
		if p.cond == nil {
			p.cond = sync.NewCond(&p.mu)
		}
		p.cond.Wait()
	}

	return nil, errors.New("SockPool: error")
}

// release release a connection, will be call when it expires or when a new connection fails.
func (p *SockPool) release() {
	p.curConns -= 1
	if p.cond != nil {
		p.cond.Signal()
	}
}

// Put connection back into the pool.
// Forced close connection when forceClose is true.
func (p *SockPool) Put(c net.Conn, forceClose bool) {
	p.mu.Lock()
	if c == nil {
		p.release()
		p.mu.Unlock()
		return
	}

	isPut := false
	if !forceClose && !p.closed {
		p.idleConns.PushFront(IdleConn{t: nowFunc(), c: c})
		if p.maxIdle > 0 && p.idleConns.Len() > p.maxIdle {
			// delete the connection when idle number exceeds the maximum idle number.
			c = p.idleConns.Remove(p.idleConns.Back()).(IdleConn).c
		} else {
			c = nil
			isPut = true
		}
	}

	// back into pool
	if isPut {
		if p.cond != nil {
			p.cond.Signal()
		}
		p.mu.Unlock()
		return
	}

	p.release()
	p.mu.Unlock()

	if c != nil {
		c.Close()
	}

	return
}

// Close close connection pool.
func (p *SockPool) Close() {
	p.mu.Lock()
	idle := p.idleConns
	p.idleConns.Init()
	p.closed = true
	p.curConns -= idle.Len()
	if p.cond != nil {
		p.cond.Broadcast()
	}
	p.mu.Unlock()
	for e := idle.Front(); e != nil; e = e.Next() {
		e.Value.(IdleConn).c.Close()
	}
	return
}
