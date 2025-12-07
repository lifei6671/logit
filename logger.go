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
		zap.AddCallerSkip(1),
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
	buf := getBuf(ctx)
	if buf == nil {
		return
	}

	buf.mu.Lock()
	defer buf.mu.Unlock()

	final := make([]zap.Field, 0)
	fkv := make(map[string]zap.Field, len(buf.normalFields)+len(buf.levelFields))

	// 1）保证元数据顺序
	for _, k := range buf.metaOrder {
		if f, ok := buf.metaFields[k]; ok {
			final = append(final, f)
			fkv[k] = f
		}
	}

	if lvl == zap.InfoLevel {
		// 2）普通字段顺序
		for _, k := range buf.normalOrder {
			if f, ok := buf.normalFields[k]; ok {
				if _, ok := fkv[k]; !ok {
					// 这里不能覆盖元数据字段
					final = append(final, f)
				}
			}
		}
	}

	// 3）level 字段严格保持顺序
	for _, k := range buf.levelOrder[lvl] {
		if f, ok := buf.levelFields[lvl][k]; ok {
			if _, ok := fkv[k]; !ok {
				// 这里不能覆盖元数据字段
				final = append(final, f)
			}
		}
	}
	// 4）最后补充字段
	for _, field := range fields {
		if _, ok := fkv[field.Key]; !ok {
			final = append(final, field)
		}
	}
	l.Logger.Log(lvl, msg, final...)
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
