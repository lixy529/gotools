package config

import (
	"bufio"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	Line_Char = '\\'
	Line_Str  = "\\"
	Inc_Str   = "include "
)

type Config struct {
	filePath    string                       // Configuration file path
	includeFile []string                     // Including configuration files
	configList  map[string]map[string]string // Configuration file content
}

// NewConfig instance object.
func NewConfig(filePath string) (*Config, error) {
	cfg := &Config{filePath: filePath, configList: make(map[string]map[string]string)}
	err := cfg.readConfig()

	return cfg, err
}

// readConfig read configuration.
func (c *Config) readConfig() error {
	err := c.parseOne(c.filePath)
	if err != nil {
		return err
	}

	// Including configuration files
	for _, file := range c.includeFile {
		if path.IsAbs(file) {
			err = c.parseOne(file)
			if err != nil {
				return err
			}
		}

		p, _ := path.Split(c.filePath)
		f := path.Join(p, file)
		err = c.parseOne(f)
		if err != nil {
			return err
		}
	}

	return nil
}

// parseOne read a configuration file.
func (c *Config) parseOne(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer file.Close()
	var section string
	buf := bufio.NewReader(file)
	var realLine string

	for {
		l, err := buf.ReadString('\n')
		line := strings.TrimSpace(l)
		n := len(line)
		if err != nil {
			if err != io.EOF {
				return err
			}

			if n == 0 {
				break
			}
		}

		if n > 0 && line[n-1] == Line_Char {
			realLine += strings.TrimSpace(strings.TrimRight(line, Line_Str))
			continue
		} else if len(realLine) > 0 {
			realLine += strings.TrimSpace(line)
		} else {
			realLine = line
		}

		n = len(realLine)
		switch {
		case n == 0:
		case string(realLine[0]) == "#": // comment
			realLine = ""
		case realLine[0] == '[' && realLine[len(realLine)-1] == ']':
			section = strings.ToUpper(strings.TrimSpace(realLine[1 : len(realLine)-1]))
			c.configList[section] = make(map[string]string)
			realLine = ""
		case n > 8 && realLine[0:8] == Inc_Str: // include
			f := realLine[8:]
			c.includeFile = append(c.includeFile, f)
			realLine = ""

		default:
			tmpLine := realLine
			if i := strings.IndexAny(realLine, "#"); i > 0 {
				tmpLine = realLine[0:i]
			}
			realLine = ""
			i := strings.IndexAny(tmpLine, "=")
			if i < 1 {
				continue
			}
			key := strings.ToUpper(strings.TrimSpace(tmpLine[0:i]))
			value := strings.TrimSpace(tmpLine[i+1 : len(tmpLine)])
			c.configList[section][key] = value
		}
	}

	return nil
}

// getValue get value by key.
func (c *Config) getValue(section, key string) (string, bool) {
	if mapSec, ok := c.configList[strings.ToUpper(section)]; ok {
		if val, ok := mapSec[strings.ToUpper(key)]; ok {
			return val, true
		}
	}

	return "", false
}

// GetString get string value by key.
// Returns the default value if the key value does not exist.
func (c *Config) GetString(section, key string, def ...string) string {
	if val, ok := c.getValue(section, key); ok {
		return val
	}

	def = append(def, "")
	return def[0]
}

// GetBool get bool value by key.
// Returns the default value if the key value does not exist.
func (c *Config) GetBool(section, key string, def ...bool) bool {
	if val, ok := c.getValue(section, key); ok {
		switch strings.ToUpper(val) {
		case "1", "T", "TRUE", "YES", "Y", "ON":
			return true
		default:
			return false
		}
	}

	def = append(def, false)
	return def[0]
}

// GetInt get int value by key.
// Returns the default value if the key value does not exist.
func (c *Config) GetInt(section, key string, def ...int) int {
	if val, ok := c.getValue(section, key); ok {
		if val, err := strconv.Atoi(val); err == nil {
			return val
		} else {
			def = append(def, 0)
			return def[0]
		}
	}

	def = append(def, 0)
	return def[0]
}

// GetInt32 get int32 value by key.
// Returns the default value if the key value does not exist.
func (c *Config) GetInt32(section, key string, def ...int32) int32 {
	if val, ok := c.getValue(section, key); ok {
		if val, err := strconv.ParseInt(val, 10, 64); err == nil {
			return int32(val)
		} else {
			def = append(def, 0)
			return def[0]
		}
	}

	def = append(def, 0)
	return def[0]
}

// GetInt64 get int64 value by key.
// Returns the default value if the key value does not exist.
func (c *Config) GetInt64(section, key string, def ...int64) int64 {
	if val, ok := c.getValue(section, key); ok {
		if val, err := strconv.ParseInt(val, 10, 64); err == nil {
			return val
		} else {
			def = append(def, 0)
			return def[0]
		}
	}

	def = append(def, 0)
	return def[0]
}

// GetFloat64 get float64 value by key.
// Returns the default value if the key value does not exist.
func (c *Config) GetFloat64(section, key string, def ...float64) float64 {
	if val, ok := c.getValue(section, key); ok {
		if val, err := strconv.ParseFloat(val, 64); err == nil {
			return val
		} else {
			def = append(def, 0.00)
			return def[0]
		}
	}

	def = append(def, 0.00)
	return def[0]
}

// SetValue set a key value.
// Create a new one if the key value does not exist.
// Update value if the key value is exist.
func (c *Config) SetValue(section, key, value string) {
	section = strings.ToUpper(section)
	key = strings.ToUpper(key)
	_, ok := c.configList[section]
	if !ok {
		c.configList[section] = make(map[string]string)
		c.configList[section][key] = value
		return
	}

	c.configList[section][key] = value
}

// GetSec return all configurations under section.
func (c *Config) GetSec(section string) (map[string]string, bool) {
	mapSec, ok := c.configList[strings.ToUpper(section)]
	if ok {
		return mapSec, true
	}

	return mapSec, false
}

// GetSecs return all section names
func (c *Config) GetSecs() []string {
	var secs []string
	for sec, _ := range c.configList {
		secs = append(secs, sec)
	}

	return secs
}
