package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog"
)

var (
	log      zerolog.Logger
	logFile  *os.File
	initOnce sync.Once
)

// Init initializes the zerolog logger with both file and console outputs.
func Init(isProduction bool, logFilePath string, level zerolog.Level) {
	initOnce.Do(func() {
		// ✅ Tạo thư mục cha nếu chưa tồn tại
		dir := filepath.Dir(logFilePath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			panic("Failed to create log directory: " + err.Error())
		}

		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic("Failed to open log file: " + err.Error())
		}
		logFile = file

		var writers []io.Writer
		writers = append(writers, logFile)

		if isProduction {
			writers = append(writers, os.Stdout)
			zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		} else {
			consoleWriter := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
				w.Out = os.Stdout
				w.TimeFormat = "2006-01-02 15:04:05"
			})
			writers = append(writers, consoleWriter)
		}

		multi := io.MultiWriter(writers...)
		log = zerolog.New(multi).
			Level(level).
			With().
			Timestamp().
			Caller().
			Logger()
	})
}

// internal helper to build structured fields
func applyFields(event *zerolog.Event, fields ...any) *zerolog.Event {
	for i := 0; i < len(fields)-1; i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		event = event.Interface(key, fields[i+1])
	}
	return event
}

// Info logs an info-level message.
func Info(msg string, fields ...any) {
	applyFields(log.Info(), fields...).Msg(msg)
}

// Error logs an error-level message.
func Error(msg string, fields ...any) {
	applyFields(log.Error(), fields...).Msg(msg)
}

// Debug logs a debug-level message.
func Debug(msg string, fields ...any) {
	applyFields(log.Debug(), fields...).Msg(msg)
}

// Warn logs a warn-level message.
func Warn(msg string, fields ...any) {
	applyFields(log.Warn(), fields...).Msg(msg)
}

// Fatal logs a fatal-level message and exits.
func Fatal(msg string, fields ...any) {
	applyFields(log.Fatal(), fields...).Msg(msg)
}

// Close closes the log file if open.
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// L returns the underlying zerolog.Logger
func L() zerolog.Logger {
	return log
}
