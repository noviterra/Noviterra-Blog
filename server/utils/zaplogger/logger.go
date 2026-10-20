package zaplogger

import "go.uber.org/zap"

type Logger struct {
	logger *zap.Logger
}

func NewLogger() *Logger {
	return &Logger{
		logger: generateLogger(),
	}
}

func (z *Logger) Debug(msg string, fields ...zap.Field) {
	z.logger.Debug(msg, fields...)
}

func (z *Logger) Info(msg string, fields ...zap.Field) {
	z.logger.Info(msg, fields...)
}

func (z *Logger) Warn(msg string, fields ...zap.Field) {
	z.logger.Warn(msg, fields...)
}

func (z *Logger) Error(msg string, fields ...zap.Field) {
	z.logger.Error(msg, fields...)
}

func (z *Logger) Panic(msg string, fields ...zap.Field) {
	z.logger.Panic(msg, fields...)
}

func (z *Logger) Fatal(msg string, fields ...zap.Field) {
	z.logger.Fatal(msg, fields...)
}
