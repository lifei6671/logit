package logit

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapHandler implements slog.Handler
type ZapHandler struct {
	logger *Logger
	fields []zap.Field
}

func NewZapHandler(logger *Logger) *ZapHandler {
	return &ZapHandler{logger: logger}
}

func (h *ZapHandler) Enabled(_ context.Context, level slog.Level) bool {
	switch level {
	case slog.LevelDebug:
		return h.logger.Core().Enabled(zap.DebugLevel)
	case slog.LevelInfo:
		return h.logger.Core().Enabled(zap.InfoLevel)
	case slog.LevelWarn:
		return h.logger.Core().Enabled(zap.WarnLevel)
	case slog.LevelError:
		return h.logger.Core().Enabled(zap.ErrorLevel)
	}
	return true
}

func (h *ZapHandler) Handle(ctx context.Context, record slog.Record) error {
	buf := getBuf(ctx)

	fields := make([]zap.Field, 0, len(h.fields)+record.NumAttrs())

	if buf != nil {
		fkv := make(map[string]zap.Field, len(buf.normalFields)+len(buf.levelFields))

		// 1）保证元数据顺序
		for _, k := range buf.metaOrder {
			if f, ok := buf.metaFields[k]; ok {
				fields = append(fields, f)
				fkv[k] = f
			}
		}
		// 这里将 slog 的日志级别转换为 zap 的级别
		lvl := levelToZapLevel(record.Level)

		if lvl == zap.InfoLevel {
			// 2）普通字段顺序
			for _, k := range buf.normalOrder {
				if f, ok := buf.normalFields[k]; ok {
					if _, ok := fkv[k]; !ok {
						// 这里不能覆盖元数据字段
						fields = append(fields, f)
					}
				}
			}
		}

		// 3）level 字段严格保持顺序
		for _, k := range buf.levelOrder[lvl] {
			if f, ok := buf.levelFields[lvl][k]; ok {
				if _, ok := fkv[k]; !ok {
					// 这里不能覆盖元数据字段
					fields = append(fields, f)
				}
			}
		}
		// 4）最后补充字段
		for _, field := range fields {
			if _, ok := fkv[field.Key]; !ok {
				fields = append(fields, field)
			}
		}
	}
	// existing fields
	fields = append(fields, h.fields...)

	// add record attrs
	record.Attrs(func(a slog.Attr) bool {
		fields = append(fields, zapAny(a))
		return true
	})

	entry := h.logger.WithOptions()

	// source information
	if record.PC != 0 {
		fn := runtime.FuncForPC(record.PC)
		file, line := fn.FileLine(record.PC)
		fields = append(fields, zap.String("caller", file+":"+toString(line)))
	}

	switch record.Level {
	case slog.LevelDebug:
		entry.Debug(record.Message, fields...)
	case slog.LevelInfo:
		entry.Info(record.Message, fields...)
	case slog.LevelWarn:
		entry.Warn(record.Message, fields...)
	case slog.LevelError:
		entry.Error(record.Message, fields...)
	default:
		entry.Info(record.Message, fields...)
	}

	return nil
}

func (h *ZapHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	child := *h // shallow copy
	for _, a := range attrs {
		child.fields = append(child.fields, zapAny(a))
	}
	return &child
}

func (h *ZapHandler) WithGroup(_ string) slog.Handler {
	// you can implement grouped prefix here if needed
	return h
}

// convert slog.Attr to zap.Field
func zapAny(a slog.Attr) zap.Field {
	switch v := a.Value.Any().(type) {
	case string:
		return zap.String(a.Key, v)
	case int:
		return zap.Int(a.Key, v)
	case int64:
		return zap.Int64(a.Key, v)
	case float64:
		return zap.Float64(a.Key, v)
	case bool:
		return zap.Bool(a.Key, v)
	default:
		return zap.Any(a.Key, v)
	}
}

func toString(i int) string {
	return fmt.Sprintf("%d", i)
}

func levelToZapLevel(level slog.Level) zapcore.Level {
	switch level {
	case slog.LevelDebug:
		return zap.DebugLevel
	case slog.LevelInfo:
		return zap.InfoLevel
	case slog.LevelWarn:
		return zap.WarnLevel
	case slog.LevelError:
		return zap.ErrorLevel

	default:
		return zap.InfoLevel
	}
}

// NewSlogLogger 将 zap 日志组件包装为 slog 内置日志组件
func NewSlogLogger(core zapcore.Core, options ...zap.Option) *slog.Logger {
	logger := zap.New(core).WithOptions(options...)
	handler := NewZapHandler(&Logger{
		Logger: logger,
	})
	return slog.New(handler)
}
