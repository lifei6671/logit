package logit

import (
	"context"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	DefaultLogger *Logger
	once          sync.Once
)

type Config struct {
	Filename   string
	MaxSize    int // MB
	MaxBackups int
	MaxAge     int // days
	Compress   bool
	Level      string // debug, info, warn, error
	ToStdout   bool
	Encoder    zapcore.Encoder
}

type Logger struct {
	*zap.Logger
}

// InitLogger 初始化全局日志字段
func InitLogger(config Config) {
	once.Do(func() {
		DefaultLogger = New(config)
	})
}

// New 初始化日志对象
func New(cfg Config) *Logger {
	writeSyncer := getWriter(cfg)
	encoder := cfg.Encoder
	if encoder == nil {
		encoder = getEncoder()
	}

	level := parseLevel(cfg.Level)

	core := zapcore.NewCore(encoder, writeSyncer, level)
	if cfg.ToStdout {
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.AddSync(os.Stdout),
			level,
		)
		core = zapcore.NewTee(core, consoleCore)
	}

	l := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(2),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	return &Logger{l}
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.Output(ctx, zap.DebugLevel, msg, fields...)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.Output(ctx, zap.InfoLevel, msg, fields...)
}

func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.Output(ctx, zap.WarnLevel, msg, fields...)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.Output(ctx, zap.ErrorLevel, msg, fields...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	l.Output(ctx, zap.FatalLevel, msg, fields...)
}

func (l *Logger) Panic(ctx context.Context, msg string, fields ...zap.Field) {
	l.Output(ctx, zap.PanicLevel, msg, fields...)
}

// Output 日志刷入磁盘
func (l *Logger) Output(ctx context.Context, lvl zapcore.Level, msg string, fields ...zap.Field) {
	final := allFields(ctx, lvl, fields...)
	l.Logger.Log(lvl, msg, final...)
}

// Flush 将各个级别的日志统一写入磁盘
func (l *Logger) Flush(ctx context.Context) {
	buf := getBuf(ctx)
	if buf == nil {
		return
	}

	buf.mu.Lock()
	defer buf.mu.Unlock()

	for lvl := range buf.levelOrder {
		l.Output(ctx, lvl, "")
	}
}

func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

func getWriter(cfg Config) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	return zapcore.AddSync(lumberJackLogger)
}

func getEncoder() zapcore.Encoder {
	cfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	return zapcore.NewJSONEncoder(cfg)
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.DateTime)) // 2025-01-08 12:22:51
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zap.DebugLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}
