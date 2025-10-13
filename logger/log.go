package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

func InitLogger() {
	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	)

	logger := zap.New(core, zap.AddCaller())
	Log = logger.Sugar()
}
