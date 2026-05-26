package telemetry

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
)

type Logger struct {
	level string
	base  *log.Logger
}

func NewLogger(level string) *Logger {
	return &Logger{
		level: strings.ToLower(level),
		base:  log.New(os.Stdout, "", 0),
	}
}

func (l *Logger) Debug(msg string, keyvals ...any) {
	if l.level == "debug" {
		l.write("debug", msg, keyvals...)
	}
}

func (l *Logger) Info(msg string, keyvals ...any) {
	l.write("info", msg, keyvals...)
}

func (l *Logger) Warn(msg string, keyvals ...any) {
	if l.level == "debug" || l.level == "info" || l.level == "warn" {
		l.write("warn", msg, keyvals...)
	}
}

func (l *Logger) Error(msg string, keyvals ...any) {
	l.write("error", msg, keyvals...)
}

func (l *Logger) write(level string, msg string, keyvals ...any) {
	record := map[string]any{
		"ts":    time.Now().UTC().Format(time.RFC3339Nano),
		"level": level,
		"msg":   msg,
	}

	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			continue
		}

		if i+1 < len(keyvals) {
			record[key] = keyvals[i+1]
		}
	}

	raw, err := json.Marshal(record)
	if err != nil {
		l.base.Printf(`{"level":"error","msg":"failed to marshal log record","error":%q}`, err.Error())
		return
	}

	l.base.Println(string(raw))
}
