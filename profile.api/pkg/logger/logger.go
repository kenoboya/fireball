package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	once         sync.Once
)

func InitLogger() {
	once.Do(func() {
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "timestamp"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

		config := zap.Config{
			Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
			Development:       false,
			DisableCaller:     false,
			DisableStacktrace: false,
			Sampling:          nil,
			Encoding:          "json",
			EncoderConfig:     encoderCfg,
			OutputPaths: []string{
				"stderr",
			},
			ErrorOutputPaths: []string{
				"stderr",
			},
			InitialFields: map[string]interface{}{
				"pid": os.Getpid(),
			},
		}
		globalLogger = zap.Must(config.Build())
	})
}
func GetLogger() *zap.Logger {
	if globalLogger == nil {
		InitLogger()
	}
	return globalLogger
}

func Debug(args ...interface{}) {
	globalLogger.Sugar().Debug(args...)
}
func Debugf(format string, args ...interface{}) {
	globalLogger.Sugar().Debugf(format, args...)
}
func Info(args ...interface{}) {
	globalLogger.Sugar().Info(args...)
}
func Infof(format string, args ...interface{}) {
	globalLogger.Sugar().Infof(format, args...)
}
func Warn(args ...interface{}) {
	globalLogger.Sugar().Warn(args...)
}
func Warnf(format string, args ...interface{}) {
	globalLogger.Sugar().Warnf(format, args...)
}
func Error(args ...interface{}) {
	globalLogger.Sugar().Error(args...)
}
func Errorf(format string, args ...interface{}) {
	globalLogger.Sugar().Errorf(format, args...)
}
func Panic(args ...interface{}) {
	globalLogger.Sugar().Panic(args...)
}
func Panicf(format string, args ...interface{}) {
	globalLogger.Sugar().Panicf(format, args...)
}
func Fatal(args ...interface{}) {
	globalLogger.Sugar().Fatal(args...)
}
func Fatalf(format string, args ...interface{}) {
	globalLogger.Sugar().Fatalf(format, args...)
}
func Log(lvl zapcore.Level, args ...interface{}) {
	globalLogger.Sugar().Log(lvl, args...)
}
func Logf(lvl zapcore.Level, format string, args ...interface{}) {
	globalLogger.Sugar().Logf(lvl, format, args...)
}
