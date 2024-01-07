// Package logger implements an interface behind which a third party, levelled
// logger can sit. This abstraction allows us to readily swap out the logger
// used and to pass it down throughout the spok program without changing
// the logger being a massive task.
//
// Spok's logging needs are fairly basic, it really only needs DEBUG level logs
// for the --debug flag.
package logger

import "go.uber.org/zap"

// Logger is the interface behind which a debug logger can sit.
type Logger interface {
	// Sync flushes the logs to stderr
	Sync() error
	// Debug outputs a debug level log line
	Debug(format string, args ...any)
}

// ZapLogger is a Logger that uses zap under the hood.
type ZapLogger struct {
	inner *zap.SugaredLogger
}

// NewZapLogger builds and returns a ZapLogger.
func NewZapLogger(verbose bool) (*ZapLogger, error) {
	level := zap.InfoLevel
	if verbose {
		level = zap.DebugLevel
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.DisableCaller = true
	logger, err := cfg.Build(zap.IncreaseLevel(level))
	if err != nil {
		return nil, err
	}
	sugar := logger.Sugar()

	return &ZapLogger{inner: sugar}, nil
}

// Sync flushes the logs.
func (z *ZapLogger) Sync() error {
	return z.inner.Sync()
}

// Debug outputs a debug level log line, a newline is automatically added.
func (z *ZapLogger) Debug(format string, args ...any) {
	z.inner.Debugf(format, args...)
}
