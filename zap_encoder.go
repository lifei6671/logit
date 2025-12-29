package logit

import "go.uber.org/zap/zapcore"

type EncoderBuilder func() zapcore.Encoder

// DefaultEncoder 默认使用 JSON 编码器
func DefaultEncoder() zapcore.Encoder {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		MessageKey:     "msg",
		LevelKey:       "level",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	return zapcore.NewJSONEncoder(encoderCfg)
}

// NewJSONEncoder json 编码器
func NewJSONEncoder() EncoderBuilder {
	return func() zapcore.Encoder {
		encoderCfg := zapcore.EncoderConfig{
			TimeKey:        "ts",
			MessageKey:     "msg",
			LevelKey:       "level",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		}
		return zapcore.NewJSONEncoder(encoderCfg)
	}
}

// NewConsoleEncoder 控制台编码器
func NewConsoleEncoder() EncoderBuilder {
	return func() zapcore.Encoder {
		encoderCfg := zapcore.EncoderConfig{
			TimeKey:        "ts",
			MessageKey:     "msg",
			LevelKey:       "level",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		}
		return zapcore.NewConsoleEncoder(encoderCfg)
	}
}
