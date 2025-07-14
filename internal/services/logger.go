package services

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"flight-booking/internal/config"
	"flight-booking/internal/interfaces"
)

// ZapLogger implements the Logger interface using zap
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new zap logger
func NewZapLogger(config config.LogConfig) (interfaces.Logger, error) {
	// Configure log level
	var level zapcore.Level
	switch config.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Configure encoder
	var encoderConfig zapcore.EncoderConfig
	if config.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	// Configure time format
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create encoder
	var encoder zapcore.Encoder
	if config.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &ZapLogger{logger: logger}, nil
}

// Debug logs a debug message
func (l *ZapLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Debug(msg, l.convertFields(fields...)...)
}

// Info logs an info message
func (l *ZapLogger) Info(msg string, fields ...interface{}) {
	l.logger.Info(msg, l.convertFields(fields...)...)
}

// Warn logs a warning message
func (l *ZapLogger) Warn(msg string, fields ...interface{}) {
	l.logger.Warn(msg, l.convertFields(fields...)...)
}

// Error logs an error message
func (l *ZapLogger) Error(msg string, fields ...interface{}) {
	l.logger.Error(msg, l.convertFields(fields...)...)
}

// With creates a new logger with additional fields
func (l *ZapLogger) With(fields ...interface{}) interfaces.Logger {
	return &ZapLogger{
		logger: l.logger.With(l.convertFields(fields...)...),
	}
}

// convertFields converts key-value pairs to zap fields
func (l *ZapLogger) convertFields(fields ...interface{}) []zap.Field {
	if len(fields)%2 != 0 {
		// If odd number of arguments, treat last as a field without value
		fields = append(fields, nil)
	}

	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}

		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}

	return zapFields
}

// Sync flushes any buffered log entries
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

// Close closes the logger
func (l *ZapLogger) Close() error {
	return l.logger.Sync()
}
