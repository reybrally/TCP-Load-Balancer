package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ANSI Color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"

	ColorBoldRed    = "\033[1;31m"
	ColorBoldGreen  = "\033[1;32m"
	ColorBoldYellow = "\033[1;33m"
	ColorBoldBlue   = "\033[1;34m"
	ColorBoldPurple = "\033[1;35m"
	ColorBoldCyan   = "\033[1;36m"
)

type Logger struct {
	*zap.Logger
}

func New(env string) *Logger {
	var logger *zap.Logger

	if env == "development" {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = customLevelEncoder
		config.EncoderConfig.EncodeTime = customTimeEncoder
		config.EncoderConfig.EncodeCaller = customCallerEncoder
		config.EncoderConfig.ConsoleSeparator = " "

		logger = zap.Must(config.Build())
	} else {
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "timestamp"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

		config := zap.Config{
			Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
			Development:       false,
			DisableCaller:     false,
			DisableStacktrace: false,
			Encoding:          "json",
			EncoderConfig:     encoderCfg,
			OutputPaths:       []string{"stdout"},
			ErrorOutputPaths:  []string{"stdout"},
			InitialFields: map[string]interface{}{
				"pid": os.Getpid(),
			},
		}

		logger = zap.Must(config.Build())
	}

	return &Logger{logger}
}

func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var levelStr string
	switch level {
	case zapcore.DebugLevel:
		levelStr = fmt.Sprintf("%s DEBUG%s", ColorGray, ColorReset)
	case zapcore.InfoLevel:
		levelStr = fmt.Sprintf("%s INFO %s", ColorBoldGreen, ColorReset)
	case zapcore.WarnLevel:
		levelStr = fmt.Sprintf("%s️  WARN %s", ColorBoldYellow, ColorReset)
	case zapcore.ErrorLevel:
		levelStr = fmt.Sprintf("%s ERROR%s", ColorBoldRed, ColorReset)
	case zapcore.DPanicLevel:
		levelStr = fmt.Sprintf("%s PANIC%s", ColorBoldPurple, ColorReset)
	case zapcore.PanicLevel:
		levelStr = fmt.Sprintf("%s PANIC%s", ColorBoldPurple, ColorReset)
	case zapcore.FatalLevel:
		levelStr = fmt.Sprintf("%s FATAL%s", ColorBoldRed, ColorReset)
	default:
		levelStr = level.CapitalString()
	}
	enc.AppendString(levelStr)
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	timeStr := fmt.Sprintf("%s[%s]%s", ColorCyan, t.Format("15:04:05.000"), ColorReset)
	enc.AppendString(timeStr)
}

func customCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	callerStr := fmt.Sprintf("%s📍 %s%s", ColorBlue, caller.TrimmedPath(), ColorReset)
	enc.AppendString(callerStr)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.Sugar().Infof(template, args...)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.Sugar().Warnf(template, args...)
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.Sugar().Debugf(template, args...)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.Sugar().Errorf(template, args...)
}

func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.Sugar().Fatalf(template, args...)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}

func (l *Logger) PrintBanner(version, port string) {
	banner := fmt.Sprintf(`
%s╔═══════════════════════════════════════════════════════╗
║                                                         ║
║          %sTCP LOAD BALANCER%s                          ║
║                                                         ║
║  %s⚡ Version:%s %-10s                                   ║
║  %s🚀 Port:%s    %-10s                                  ║
║  %s🔥 Status:%s   %sRUNNING%s                           ║
║                                                         ║
╚═════════════════════════════════════════════════════════╝%s
`, ColorBoldCyan, ColorBoldPurple, ColorBoldCyan,
		ColorYellow, ColorWhite, version,
		ColorYellow, ColorWhite, port,
		ColorYellow, ColorWhite, ColorBoldGreen, ColorWhite,
		ColorReset)

	fmt.Println(banner)
}
