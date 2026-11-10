package zaplogger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// JSON输出格式
func jsonEncoder() zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncodeCaller = zapcore.ShortCallerEncoder

	return zapcore.NewJSONEncoder(config)
}

// 输出配置--文件
func fileWriterSyncer() zapcore.WriteSyncer {
	lumber := &lumberjack.Logger{
		Filename:   "D:\\GoProject\\IT\\Noviterra\\log\\log1.log",
		MaxSize:    100,
		MaxAge:     7,
		MaxBackups: 5,
		Compress:   true,
	}

	return zapcore.AddSync(lumber)
}

// 美观输出 到Terminal
func beautifulEncoder() zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncodeCaller = zapcore.ShortCallerEncoder

	return zapcore.NewConsoleEncoder(config)
}

// 输出配置--console
func consoleWriterSyncer() zapcore.WriteSyncer {
	return zapcore.AddSync(os.Stdout)
}

// 异步日志
func warpSyncCore(ws zapcore.WriteSyncer) *zapcore.BufferedWriteSyncer {
	syncWS := &zapcore.BufferedWriteSyncer{
		WS:            ws,
		Size:          1024 * 1024,
		FlushInterval: 5 * time.Second,
		Clock:         zapcore.DefaultClock,
	}

	return syncWS
}

func warpSampler(core zapcore.Core) zapcore.Core {
	return zapcore.NewSampler(core, time.Second, 100, 1)
}

func newCore() zapcore.Core {
	level := zapcore.InfoLevel

	// 组装文件、控制台输出的核心
	fileCore := zapcore.NewCore(jsonEncoder(), warpSyncCore(fileWriterSyncer()), level)
	consoleCore := zapcore.NewCore(beautifulEncoder(), warpSyncCore(consoleWriterSyncer()), level)

	// 日志核心
	core := zapcore.NewTee(fileCore, consoleCore)

	return core
}

// 生成日志实例
func generateLogger() *zap.Logger {
	core := newCore()
	return zap.New(
		core,
		zap.WrapCore(warpSampler),
		zap.AddCaller(),
		zap.AddCallerSkip(1), // 跳过封装 caller显示业务调用位置
	)
}
