package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log     *zap.Logger
	logFile *os.File
)

// Init initializes the zap logger with file + console output.
func Init(isProduction bool, logFilePath string, level zapcore.Level) {
	var encoderConfig zapcore.EncoderConfig
	if isProduction {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Open log file
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("Failed to open log file: " + err.Error())
	}
	logFile = file

	// File writer core (JSON)
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(file),
		level,
	)

	// Console writer core (pretty)
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Combine both cores
	core := zapcore.NewTee(fileCore, consoleCore)

	// Create logger with caller and stacktrace options
	log = zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	// Optional: make zap global
	zap.ReplaceGlobals(log)
}

// Info logs an info-level message.
func Info(msg string, fields ...zap.Field) {
	if log != nil {
		log.Info(msg, fields...)
	}
}

// Error logs an error-level message.
func Error(msg string, fields ...zap.Field) {
	if log != nil {
		log.Error(msg, fields...)
	}
}

// Debug logs a debug-level message.
func Debug(msg string, fields ...zap.Field) {
	if log != nil {
		log.Debug(msg, fields...)
	}
}

// Warn logs a warn-level message.
func Warn(msg string, fields ...zap.Field) {
	if log != nil {
		log.Warn(msg, fields...)
	}
}

// Sync flushes the logger buffers.
func Sync() {
	if log != nil {
		log.Sync()
	}
}

// Close safely closes the log file.
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// L returns the underlying *zap.Logger
func L() *zap.Logger {
	return log
}
