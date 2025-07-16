package logger

import (
	"context"
	"fmt"
	"os"

	"flight-booking/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey struct{}

type Logger interface {
	SetIntoContext(ctx context.Context) context.Context
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	With(fields ...interface{}) Logger
}

func Context(ctx context.Context) Logger {
	if logger, ok := ctx.Value(ctxKey{}).(Logger); ok {

		return logger
	}

	return &logger{logger: zap.NewNop()}
}

type logger struct {
	logger *zap.Logger
}

func New(config config.Config) (Logger, error) {
	var level zapcore.Level

	switch config.Log.Level {
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

	var encoderConfig zapcore.EncoderConfig
	if config.Log.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if config.Log.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	return &logger{
		logger: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)),
	}, nil
}

func (l *logger) SetIntoContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

func (l *logger) Debug(msg string, fields ...interface{}) {
	l.logger.Debug(msg, l.convertFields(fields...)...)
}

func (l *logger) Info(msg string, fields ...interface{}) {
	l.logger.Info(msg, l.convertFields(fields...)...)
}

func (l *logger) Warn(msg string, fields ...interface{}) {
	l.logger.Warn(msg, l.convertFields(fields...)...)
}

func (l *logger) Error(msg string, fields ...interface{}) {
	l.logger.Error(msg, l.convertFields(fields...)...)
}

func (l *logger) With(fields ...interface{}) Logger {
	return &logger{
		logger: l.logger.With(l.convertFields(fields...)...),
	}
}

func (l *logger) convertFields(fields ...interface{}) []zap.Field {
	if len(fields)%2 != 0 {
		fields = append(fields, nil)
	}

	zapFields := make([]zap.Field, 0, len(fields))

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

func (l *logger) Sync() error {
	err := l.logger.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync logger: %w", err)
	}

	return nil
}

func (l *logger) Close() error {
	return l.Sync()
}
