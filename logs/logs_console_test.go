package logs

import (
	"testing"
)

func TestConsole(t *testing.T) {
	l := &ConsoleLogs{}
	l.Init(`{"level":1, "showcall":true, "depth":3}`)

	l.Debug(LevelDebug, "LevelDebug")
	l.Info(LevelInfo, "LevelInfo")
	l.Warn(LevelWarn, "LevelWarn")
	l.Error(LevelError, "LevelError")
	l.Fatal(LevelFatal, "LevelFatal")

	l.Debugf("Debugf %d-%s", LevelDebug, "LevelDebug")
	l.Infof("Infof %d-%s", LevelInfo, "LevelInfo")
	l.Warnf("Warnf %d-%s", LevelWarn, "LevelWarn")
	l.Errorf("Errorf %d-%s", LevelError, "LevelError")
	l.Fatalf("Fatalf %d-%s", LevelFatal, "LevelFatal")
}
