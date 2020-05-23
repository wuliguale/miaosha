package common

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger() {
	logFile := GetLogFile()

	lumberJackLogger := &lumberjack.Logger{
		Filename: logFile,
		MaxSize: 1,
		MaxBackups: 5,
		MaxAge: 30,
		Compress: false,
	}
	writeSyncer := zapcore.AddSync(lumberJackLogger)

	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	core := zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel)

	//callSkip=1 caller skip common function
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.Fields(zap.String("project", "miaosha-demo")))
	zap.ReplaceGlobals(logger)
}


func ZapError(msg string, err error) {
	zap.L().Error(msg, zap.Error(err))
}


func zapInfo(msgs ...string) {
	for _, msg := range msgs {
		zap.L().Info(msg)
	}
}


