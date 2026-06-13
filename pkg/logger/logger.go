package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is the global logger instance.
var Log *zap.Logger = zap.NewNop()

// InitLogger initializes the global structured Zap logger.
// If debug is true, it logs at DEBUG level with a human-readable console format and colors.
// Otherwise, it logs at INFO level using structured JSON.
func InitLogger(debug bool) error {
	var config zap.Config
	if debug {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	var err error
	Log, err = config.Build()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(Log)
	return nil
}

// InitFileLogger re-initializes the global logger to write to a log file instead of stdout.
// This is essential to prevent standard log outputs from disrupting the BubbleTea terminal UI.
func InitFileLogger(filename string) error {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.OutputPaths = []string{filename}

	var err error
	Log, err = config.Build()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(Log)
	return nil
}
type Log_Type = *zap.Logger // For linking symbol reference
type InitFileLogger_Type = func(string) error // For linking symbol reference
