package logit

import (
	"fmt"
	"io"
	"strings"

	"go.uber.org/zap/zapcore"
)

// ZapDispatch 规则：某类日志写入某个文件后缀对应的文件
type ZapDispatch struct {
	// FileSuffix 日志文件后缀，例如：
	// ""   表示默认 notice 文件，例如: app.log.2025010110
	// ".wf" 表示写入 warn/fatal 文件，例如: app.log.wf.2025010110
	FileSuffix string

	// Levels 指定要写入该文件的日志级别
	// 若为空，则视为无效配置。
	Levels []zapcore.Level
	// 自定义编码器
	EncoderBuilder EncoderBuilder
}

type WriterBuilder func(ruleName, filename string, opts ...ZapWriterOptions) (zapcore.WriteSyncer, error)

func DefaultWriterBuild(ruleName, filename string, opts ...ZapWriterOptions) (zapcore.WriteSyncer, error) {
	return BuildZapWriteSyncer(ruleName, filename, opts...)
}

type EncoderBuilder func() zapcore.Encoder

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

type CloseFunc func()

func BuildDefaultZapCore(
	ruleName string,
	filename string,
	dispatchRule []ZapDispatch,
	opts ...ZapWriterOptions) (zapcore.Core, CloseFunc, error) {
	return BuildDispatchCore(ruleName, filename, dispatchRule, DefaultWriterBuild, DefaultEncoder, opts...)
}
func BuildDispatchCore(
	ruleName string,
	filename string,
	dispatchRules []ZapDispatch,
	writerBuilder WriterBuilder,
	encoderBuilder EncoderBuilder,
	opts ...ZapWriterOptions,
) (zapcore.Core, CloseFunc, error) {
	var cores []zapcore.Core

	// 记录所有创建出来的 writer，方便退出时 Close/Sync
	var closers []zapcore.WriteSyncer

	for _, rule := range dispatchRules {
		if len(rule.Levels) == 0 {
			continue
		}

		file := buildDispatchFilename(filename, rule.FileSuffix)

		if writerBuilder == nil {
			writerBuilder = DefaultWriterBuild
		}
		ws, err := writerBuilder(ruleName, file, opts...)
		if err != nil {
			return nil, nil, err
		}
		// 优先使用独立编码器
		if rule.EncoderBuilder != nil {
			encoderBuilder = rule.EncoderBuilder
		}
		// 如果没传入则使用默认编码器
		if encoderBuilder == nil {
			encoderBuilder = DefaultEncoder
		}
		encoder := encoderBuilder()

		levelFilter := newLevelFilter(rule.Levels)

		core := zapcore.NewCore(
			encoder,
			ws,
			levelFilter, // 核心过滤器
		)

		cores = append(cores, core)
		closers = append(closers, ws)
	}
	if len(cores) == 0 {
		return nil, nil, fmt.Errorf("no valid dispatch rules, cores is empty")
	}
	dispatchCore := zapcore.NewTee(cores...)

	return dispatchCore, func() {
		seen := make(map[zapcore.WriteSyncer]struct{}, len(closers))
		for _, w := range closers {
			if _, ok := seen[w]; ok {
				continue
			}
			seen[w] = struct{}{}

			_ = w.Sync()
			if closer, ok := w.(io.Closer); ok {
				_ = closer.Close()
			}
		}
	}, nil
}

type levelFilter struct {
	all map[zapcore.Level]struct{}
}

func newLevelFilter(levels []zapcore.Level) zapcore.LevelEnabler {
	m := make(map[zapcore.Level]struct{}, len(levels))
	for _, lv := range levels {
		m[lv] = struct{}{}
	}
	return &levelFilter{all: m}
}

func (f *levelFilter) Enabled(l zapcore.Level) bool {
	_, ok := f.all[l]
	return ok
}

func buildDispatchFilename(baseFile string, suffix string) string {
	if suffix == "" {
		return baseFile
	}
	if !strings.HasPrefix(suffix, ".") {
		suffix = "." + suffix
	}

	// 原 baseFile: service.log
	// 新 file: service.log.wf
	// rotate时再跟你的生成器组合成 file.2025010110
	return baseFile + suffix
}
