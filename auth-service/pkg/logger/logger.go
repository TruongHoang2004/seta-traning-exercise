package logger

import (
	"io"
	"os"
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
		// Open log file
		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic("Failed to open log file: " + err.Error())
		}
		logFile = file

		// Multi writer: file + console
		var writers []io.Writer
		writers = append(writers, logFile)

		if isProduction {
			// In production, log as JSON to stdout
			writers = append(writers, os.Stdout)
			zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		} else {
			// In development, use console writer for pretty printing
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

// Info logs an info-level message.
func Info(msg string, fields ...any) {
	log.Info().Fields(fields).Msg(msg)
}

// Error logs an error-level message.
func Error(msg string, fields ...any) {
	log.Error().Fields(fields).Msg(msg)
}

// Debug logs a debug-level message.
func Debug(msg string, fields ...any) {
	log.Debug().Fields(fields).Msg(msg)
}

// Warn logs a warn-level message.
func Warn(msg string, fields ...any) {
	log.Warn().Fields(fields).Msg(msg)
}

// Fatal logs a fatal-level message and exits.
func Fatal(msg string, fields ...any) {
	log.Fatal().Fields(fields).Msg(msg)
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
